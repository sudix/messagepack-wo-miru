package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/elazarl/goproxy"
)

func TestMpmProxy(t *testing.T) {
	fmt.Println("aaa")

	// start proxy server
	client, _, s := oneShotProxy()
	defer s.Close()
	fmt.Println(s.URL)

	// start test response server
	ts := httptest.NewServer(http.HandlerFunc(dummResponseHandler))
	defer ts.Close()
	fmt.Println(ts.URL)

	res, err := client.Get(ts.URL)
	if err != nil {
		t.Error(err)
	}
	res.Body.Close()

	fmt.Sprintf("%#v¥n", res.Body)
}

func oneShotProxy() (client *http.Client, proxy *goproxy.ProxyHttpServer, s *httptest.Server) {
	proxy = buildProxy()
	s = httptest.NewServer(proxy)
	proxyUrl, _ := url.Parse(s.URL)
	tr := &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	client = &http.Client{Transport: tr}
	return
}

func dummResponseHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}
