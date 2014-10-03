package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"

	"log"
	"net/http"
	"net/http/httputil"
	"reflect"
	"strings"
	"unicode"

	"github.com/mgutz/ansi"
	"github.com/ugorji/go/codec"
)

var (
	listen = flag.String("listen", "localhost:1080", "listen on address")
	logp   = flag.Bool("log", false, "enable logging")
)

var (
	singleCRLF = []byte("\r\n")
	doubleCRLF = []byte("\r\n\r\n")
)

type Direction int

const (
	Response Direction = iota + 1
	Request
)

func main() {
	flag.Parse()
	proxyHandler := http.HandlerFunc(proxyHandlerFunc)
	fmt.Println("start proxy.")
	log.Fatal(http.ListenAndServe(*listen, proxyHandler))
}

func proxyHandlerFunc(w http.ResponseWriter, r *http.Request) {
	// Log if requested
	if *logp {
		log.Println(r.URL)
	}

	// We'll want to use a new client for every request.
	client := &http.Client{}

	// Tweak the request as appropriate:
	//RequestURI may not be sent to client
	//URL.Scheme must be lower-case
	r.RequestURI = ""
	r.URL.Scheme = strings.Map(unicode.ToLower, r.URL.Scheme)

	reqDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Fatal(err)
	}

	_, reqBody, err := splitEachSection(reqDump)
	if err != nil {
		log.Fatal(err)
	}

	outputBody(reqBody, Request)

	// And proxy
	resp, err := client.Do(r)
	if err != nil {
		log.Fatal(err)
	}

	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatal(err)
	}

	_, body, err := splitEachSection(dump)
	if err != nil {
		log.Fatal(err)
	}

	outputBody(body, Response)

	resp.Write(w)
}

func splitEachSection(dump []byte) (header, body []byte, err error) {
	sections := bytes.Split(dump, doubleCRLF)
	if len(sections) != 2 {
		return nil, nil, fmt.Errorf("section number error")
	}
	header = sections[0]
	body = sections[1]
	return header, body, nil
}

func outputBody(body []byte, direction Direction) error {
	color := "white"
	if direction == Request {
		color = "cyan"
		fmt.Println(ansi.Color("====[request]====>", color))
	} else {
		color = "green"
		fmt.Println(ansi.Color("<====[response]====", color))
	}
	json, err := msgPackToJson(body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	coloredJson := ansi.Color(json, color)
	fmt.Println(coloredJson)

	return nil
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
