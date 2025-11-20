package xray_pool

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

type CheckFunc func(n *Node) (time.Duration, error)

// ByPing 通过ping来判断
func ByPing(n *Node) (time.Duration, error) {
	conn, err := net.DialTimeout("ip:icmp", n.Hostname(), DefaultTimeout)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	t := time.Now()
	if err = conn.SetDeadline(time.Now().Add(DefaultTimeout)); err != nil {
		return 0, err
	}
	if _, err = conn.Write([]byte{
		8, 0, 247, 253, 0, 1, 0, 1, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0}); err != nil {
		return 0, err
	}
	buf := make([]byte, 128)
	_, err = conn.Read(buf)
	return time.Since(t), err
}

// ByTCP 通过dial来判断
func ByTCP(n *Node) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", n.Hostname(), n.Port()), DefaultTimeout)
	if err != nil {
		return 0, err
	}
	_ = conn.Close()
	return time.Since(start), nil
}

func ByGoogle(n *Node) (time.Duration, error) {
	start := time.Now()
	u, err := url.Parse(n.Proxy())
	if err != nil {
		return 0, err
	}
	h := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			Proxy: http.ProxyURL(u),
		},
		Timeout: time.Second * 10,
	}
	resp, err := h.Get("https://www.google.com")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return time.Since(start), nil
}
