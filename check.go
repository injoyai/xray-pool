package xray_pool

import (
	"net"
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

// ByDial 通过dial来判断
func ByDial(n *Node) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", n.Address(), DefaultTimeout)
	if err != nil {
		return 0, err
	}
	_ = conn.Close()
	return time.Since(start), nil
}
