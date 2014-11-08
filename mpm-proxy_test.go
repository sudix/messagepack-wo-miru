package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/elazarl/goproxy"
	"github.com/ugorji/go/codec"
)

var (
	mh codec.MsgpackHandle
)

func TestMpmProxy(t *testing.T) {
	// start test response server
	ts := httptest.NewServer(http.HandlerFunc(MockHandler))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err)
	}

	targetHost = u.Host

	// start proxy server
	client, _, s := oneShotProxy()
	defer s.Close()

	// request
	req := struct {
		Message string
		Number  int
	}{
		Message: "foo",
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
	res, err := client.Post(ts.URL, "application/x-msgpack", mbr)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Errorf("Response Error! status code = $d", res.StatusCode)
	}
	defer res.Body.Close()

	rb, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%s\n", rb)

	//-----------------

	// res, err := client.Get(ts.URL)
	// if err != nil {
	// 	t.Error(err)
	// }
	// defer res.Body.Close()

	// // var p []byte
	// // res.Body.Read(p)
	// b, _ := ioutil.ReadAll(res.Body)
	// fmt.Printf("%s\n", b)
}

func oneShotProxy() (client *http.Client, proxy *goproxy.ProxyHttpServer, s *httptest.Server) {
	proxy = buildProxy()
	s = httptest.NewServer(proxy)
	proxyUrl, _ := url.Parse(s.URL)
	tr := &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	client = &http.Client{Transport: tr}
	return
}

func MockHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-msgpack")
	res := struct {
		Name string
		Age  int
	}{
		Name: "bar",
		Age:  24,
	}
	enc := codec.NewEncoder(w, &mh)
	err := enc.Encode(res)
	if err != nil {
		panic(err)
	}
}
