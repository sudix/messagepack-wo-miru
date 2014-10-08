package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ugorji/go/codec"
)

var (
	mh codec.MsgpackHandle
)

func TestPrintMessagePackValues(t *testing.T) {
	// setting proxy
	proxy := httptest.NewServer(http.HandlerFunc(proxyHandlerFunc))
	defer proxy.Close()
	proxyUrl, err := url.Parse(proxy.URL)

	// test server
	ts := httptest.NewServer(http.HandlerFunc(Handler))
	defer ts.Close()

	// request
	req := struct {
		Message string
		Number  int
	}{
		Message: "fugafuga",
		Number:  12398432904832,
	}
	var b []byte
	enc := codec.NewEncoderBytes(&b, &mh)
	err = enc.Encode(req)
	if err != nil {
		t.Error(err)
	}
	mbr := bytes.NewReader(b)

	// post
	httpClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	res, err := httpClient.Post(ts.URL, "application/x-msgpack", mbr)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Errorf("Response Error! status code = $d", res.StatusCode)
	}
	defer res.Body.Close()

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-msgpack")
	res := struct {
		Name string
		Age  int
	}{
		Name: "hoge",
		Age:  24,
	}
	enc := codec.NewEncoder(w, &mh)
	err := enc.Encode(res)
	if err != nil {
		panic(err)
	}
}
