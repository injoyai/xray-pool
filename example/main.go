package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/injoyai/logs"
	xray_pool "github.com/injoyai/xray-pool"
	"golang.org/x/net/proxy"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

func main() {
	//s := "https://www.85la.com/wp-content/uploads/2025/11/202511094821bD8GXY.txt"
	s := "vless://270663ac-7abe-43e7-9f93-f0ee5ab4968c@[2001:41d0:701:1000::5505]:65531?encryption=none&security=none&type=grpc&authority="
	p := xray_pool.New(
		//xray_pool.WithSubscribe(s),
		xray_pool.WithNode(s),
	)
	defer p.Close()
	go func() {
		for i := 0; i < 10; i++ {
			logs.Info(p.Len())
			logs.PrintErr(p.Do(demo))
		}
	}()
	logs.Err(p.Run())
}

func demo(n *xray_pool.Node) error {
	logs.Info(n.Address())
	c, err := NewClient(n.Address())
	if err != nil {
		return err
	}
	resp, err := c.Get("https://www.google.com")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(bs))
	return nil
}

func NewClient(proxyUrl string) (*http.Client, error) {
	u, err := url.Parse(proxyUrl)
	if err != nil {
		return nil, err
	}
	dialer, err := proxy.FromURL(u, &net.Dialer{
		Timeout:   time.Second * 10,
		KeepAlive: time.Second * 10,
	})
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			},
		},
		Timeout: time.Second * 10,
	}, nil
}
