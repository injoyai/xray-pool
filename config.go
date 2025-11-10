package xray_pool

import (
	"encoding/json"
)

const (
	Vmess  = "vmess"
	Vless  = "vless"
	Trojan = "trojan"
	Socks  = "socks"
	None   = "none"
)

var (
	DefaultLog = Log{
		Access: "none",
		Error:  "none",
		Level:  "warning",
	}
	DemoInbounds = []Bound{
		{
			Port:     50000,
			Listen:   "127.0.0.1",
			Protocol: Socks,
			Settings: Settings{
				Udp: true,
			},
		},
	}
	DemoOutbounds = []Bound{
		{
			Protocol: Vmess,
			Settings: Settings{
				Vnext: []Vnext{
					{
						Address: "127.0.0.1",
						Port:    12345,
						Users: []User{
							{
								ID:      "00000000-0000-0000-0000-000000000000",
								AlterId: 0,
							},
						},
					},
				},
			},
		},
	}
)

type Config struct {
	Log       Log     `json:"log"`
	Inbounds  []Bound `json:"inbounds"`
	Outbounds []Bound `json:"outbounds"`
}

func (c *Config) String() string {
	return string(c.Bytes())
}

func (c *Config) Bytes() []byte {
	bs, _ := json.Marshal(c)
	return bs
}

type Log struct {
	Access string `json:"access"`
	Error  string `json:"error"`
	Level  string `json:"loglevel"`
}

type Bound struct {
	Listen   string   `json:"listen"`
	Port     int      `json:"port"`
	Protocol string   `json:"protocol"`
	Settings Settings `json:"settings"`
}

type Settings struct {
	Udp   bool    `json:"udp"`
	Vnext []Vnext `json:"vnext"`
}

type User struct {
	ID         string `json:"id"`
	AlterId    int    `json:"alterId"`
	Encryption string `json:"encryption"`
}

type Vnext struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Users   []User `json:"users"`
}

/*



 */
