package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"

	"github.com/codegangsta/cli"
	"github.com/elazarl/goproxy"
	"github.com/ugorji/go/codec"
)

var (
	proxy        *goproxy.ProxyHttpServer
	targetDomain string
	verbose      bool
)

func setAppInfo(app *cli.App) {
	app.Name = "mpwm - messagepack viewer"
	app.Usage = "show message pack valuew from file or http proxy"
	app.Version = "0.0.1"
}

func setFlags(app *cli.App) {
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "target-domain, t",
			Value: "localhost",
			Usage: "output target domain. eg. www.google.com, localhost:3000",
		},
		cli.BoolFlag{
			Name: "verbose",
			Usage: "The host that the proxy server" +
				" should forward requests to.",
		},
		cli.IntFlag{
			Name:  "port, p",
			Value: 8080,
			Usage: "proxy's port.",
		},
	}

}

func msgPackToJson(b []byte) (string, error) {
	// https://github.com/ugorji/go/issues/1
	var mh = codec.MsgpackHandle{RawToString: true}
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	//mh.MapType = reflect.TypeOf(result{})
	var v map[string]interface{}
	dec := codec.NewDecoderBytes(b, &mh)
	err := dec.Decode(&v)
	if err != nil {
		return "", err
	}

	j, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", nil
	}

	jsonString := fmt.Sprintf("%s", j)

	return jsonString, nil

}

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func onRequest() {
	proxy.OnRequest().
		HijackConnect(func(req *http.Request, client net.Conn, ctx *goproxy.ProxyCtx) {
		defer func() {
			if e := recover(); e != nil {
				ctx.Logf("error connecting to remote: %v", e)
				client.Write([]byte("HTTP/1.1 500 Cannot reach destination\r\n\r\n"))
			}
			client.Close()
		}()
		fmt.Println("&&&&&&&&")
		clientBuf := bufio.NewReadWriter(bufio.NewReader(client), bufio.NewWriter(client))
		remote, err := net.Dial("tcp", req.URL.Host)
		orPanic(err)
		remoteBuf := bufio.NewReadWriter(bufio.NewReader(remote), bufio.NewWriter(remote))
		for {
			req, err := http.ReadRequest(clientBuf.Reader)
			orPanic(err)
			orPanic(req.Write(remoteBuf))
			orPanic(remoteBuf.Flush())
			resp, err := http.ReadResponse(remoteBuf.Reader, req)
			orPanic(err)
			orPanic(resp.Write(clientBuf.Writer))
			orPanic(clientBuf.Flush())
		}
		fmt.Printf("%s\n", remoteBuf)
		fmt.Printf("%s\n", clientBuf)
	})

	// proxy.OnRequest(goproxy.DstHostIs(targetDomain)).DoFunc(
	// 	func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	// 		fmt.Println("======[ out ]======>")
	// 		fmt.Println("---[ cookie ]--->")
	// 		for _, c := range req.Cookies() {
	// 			fmt.Printf("%#v\n", c)
	// 		}
	// 		fmt.Println("---[ body ]--->")
	// 		body, err := ioutil.ReadAll(req.Body)
	// 		if err != nil {
	// 			fmt.Println("############")
	// 			fmt.Println(err)
	// 			return req, nil
	// 		}
	// 		json, err := msgPackToJson(body)
	// 		if err != nil {
	// 			fmt.Println("$$$$$$$$$$$$")
	// 			fmt.Println(err)
	// 			return req, nil
	// 		}
	// 		fmt.Println(json)
	// 		return req, nil
	// 	})
}

func onResponse() {
	proxy.OnResponse(goproxy.DstHostIs(targetDomain)).DoFunc(
		func(res *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			fmt.Println("<=====[ in ]=======")
			fmt.Println("<---[ cookie ]---")
			fmt.Println(res)
			for _, c := range res.Cookies() {
				fmt.Printf("%#v\n", c)
			}
			fmt.Println("<---[ body ]---")

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				fmt.Println(err)
				return res
			}
			json, err := msgPackToJson(body)
			if err != nil {
				fmt.Println(err)
				return res
			}
			fmt.Println(json)
			return res
		})
}

func main() {
	app := cli.NewApp()
	setAppInfo(app)
	setFlags(app)

	app.Action = func(c *cli.Context) {
		targetDomain = c.String("target-domain")
		verbose = c.Bool("verbose")

		proxy = goproxy.NewProxyHttpServer()
		proxy.Verbose = verbose

		onRequest()
		onResponse()

		log.Fatal(http.ListenAndServe(":8080", proxy))
	}
	fmt.Println("proxy start.")
	app.Run(os.Args)
}
