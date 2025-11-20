package main

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"

	"github.com/injoyai/conv"
	"github.com/injoyai/logs"
	xray_pool "github.com/injoyai/xray-pool"
)

func main() {
	s := "https://www.85la.com/wp-content/uploads/2025/11/202511094821bD8GXY.txt"
	//s := "vless://c37cdcff-42f1-4f09-8ce1-0df6cf7e2520@sandking.fonixapp.org:33115?encryption=none&flow=xtls-rprx-vision&security=reality&sni=yelp.com&fp=chrome&pbk=53Q1y0Vmf2zaGBBlcO1NyKFvQM1TShkJKBCNjlevpns&sid=09cb&spx=%2F&allowInsecure=1&type=tcp&headerType=none#%F0%9F%87%A6%F0%9F%87%B9%20www.85.com%20%E5%A5%A5%E5%9C%B0%E5%88%A9"
	p := xray_pool.New(
		xray_pool.WithSubscribe(s),
		//xray_pool.WithNode(s),
	)
	defer p.Close()
	go p.Run()
	<-p.Started()

	for i := 0; i < 1; i++ {
		logs.PrintErr(p.Do(demo))
	}
}

func demo(n *xray_pool.Node) (err error) {
	defer func() {
		logs.Info(n.Proxy(), conv.Select(err == nil, "success", conv.String(err)))
	}()
	c, err := NewClient(n.Proxy())
	if err != nil {
		return err
	}
	resp, err := c.Get("https://www.google.com")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
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
