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
	client, _, s := oneShotProxy()
	defer s.Close()
	res, err := client.Get("http://google.com/")
	if err != nil {
		t.Error(err)
	}
	fmt.Sprintf("%#v¥n", res)

}

func oneShotProxy() (client *http.Client, proxy *goproxy.ProxyHttpServer, s *httptest.Server) {
	proxy = buildProxy()
	s = httptest.NewServer(proxy)
	proxyUrl, _ := url.Parse(s.URL)
	tr := &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	client = &http.Client{Transport: tr}
	return
}
