package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/injoyai/logs"
)

func main() {
	u := "http://127.0.0.1:50006"
	logs.Info(u)
	c, err := NewClient(u)
	logs.PanicErr(err)
	resp, err := c.Get("https://www.google.com")
	logs.PanicErr(err)
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	logs.PanicErr(err)
	fmt.Println(string(bs))
}

func NewClient(proxyUrl string) (*http.Client, error) {
	u, err := url.Parse(proxyUrl)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			Proxy: http.ProxyURL(u),
		},
		Timeout: time.Second * 10,
	}, nil
}
