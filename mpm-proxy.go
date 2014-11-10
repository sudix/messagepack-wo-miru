package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"reflect"

	"github.com/codegangsta/cli"
	"github.com/elazarl/goproxy"
	"github.com/ugorji/go/codec"
)

var (
	targetHost string
	verbose    bool
	port       int
)

var (
	singleCRLF = []byte("\r\n")
	doubleCRLF = []byte("\r\n\r\n")
)

func setAppInfo(app *cli.App) {
	app.Name = "mpm-proxy"
	app.Usage = "proxy tool to show messagepack values in response and request."
	app.Version = "0.0.1"
}

func setFlags(app *cli.App) {
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "host, t",
			Value: "localhost",
			Usage: "output target host. eg. www.google.com, localhost:3000",
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

func onRequest(proxy *goproxy.ProxyHttpServer) {
	// proxy.OnRequest(goproxy.DstHostIs(targetHost)).DoFunc(
	proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			fmt.Println(req.URL)
			fmt.Println("------[ request ]------>")
			// fmt.Println("---[ body ]--->")
			fmt.Println(req.URL.Host)
			dump, err := httputil.DumpRequest(req, true)
			if err != nil {
				fmt.Println(err)
				return req, nil
			}

			_, body, err := splitEachSection(dump)
			if err != nil {
				fmt.Println(err)
				return req, nil
			}

			json, err := msgPackToJson(body)
			if err != nil {
				fmt.Println(err)
				return req, nil
			}
			fmt.Println(json)
			return req, nil
		})
}

func onResponse(proxy *goproxy.ProxyHttpServer) {
	// proxy.OnResponse(goproxy.DstHostIs(targetHost)).DoFunc(
	proxy.OnResponse().DoFunc(
		func(res *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			fmt.Println("<=====[ response ]=======")
			fmt.Println("<---[ body ]---")

			dump, err := httputil.DumpResponse(res, true)
			if err != nil {
				fmt.Println(err)
				return res

			}

			_, body, err := splitEachSection(dump)
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

func splitEachSection(dump []byte) (header, body []byte, err error) {
	sections := bytes.Split(dump, doubleCRLF)
	if len(sections) != 2 {
		return nil, nil, nil //fmt.Errorf("section number error")
	}
	header = sections[0]
	body = sections[1]
	return header, body, nil
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
		targetHost = c.String("host")
		verbose = c.Bool("verbose")
		port = c.Int("port")
		proxy := buildProxy()
		portString := fmt.Sprintf(":%d", port)
		log.Fatal(http.ListenAndServe(portString, proxy))
	}
	fmt.Println("proxy start.")
	app.Run(os.Args)
}
