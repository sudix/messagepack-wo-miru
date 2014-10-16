package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/codegangsta/cli"
	"github.com/elazarl/goproxy"
	"github.com/ugorji/go/codec"
)

var (
	targetDomain string
	verbose      bool
)

func setAppInfo(app *cli.App) {
	app.Name = "mpwm - messagepack viewer"
	app.Usage = "show message pack values from file or http proxy"
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

func onRequest(proxy *goproxy.ProxyHttpServer) {
	proxy.OnRequest(goproxy.DstHostIs(targetDomain)).DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			fmt.Println("======[ out ]======>")
			fmt.Println("---[ cookie ]--->")
			for _, c := range req.Cookies() {
				fmt.Printf("%#v\n", c)
			}
			fmt.Println("---[ body ]--->")
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Println("############")
				fmt.Println(err)
				return req, nil
			}
			json, err := msgPackToJson(body)
			if err != nil {
				fmt.Println("$$$$$$$$$$$$")
				fmt.Println(err)
				return req, nil
			}
			fmt.Println(json)
			return req, nil
		})
}

func onResponse(proxy *goproxy.ProxyHttpServer) {
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

func buildProxy() (proxy *goproxy.ProxyHttpServer) {
	proxy = goproxy.NewProxyHttpServer()
	proxy.Verbose = verbose
	onRequest(proxy)
	onResponse(proxy)
	return
}

func main() {
	app := cli.NewApp()
	setAppInfo(app)
	setFlags(app)
	app.Action = func(c *cli.Context) {
		targetDomain = c.String("target-domain")
		verbose = c.Bool("verbose")
		proxy := buildProxy()
		log.Fatal(http.ListenAndServe(":8080", proxy))
	}
	fmt.Println("proxy start.")
	app.Run(os.Args)
}
