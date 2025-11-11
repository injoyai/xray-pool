package xray_pool

import (
	"encoding/json"
)

const (
	Vmess  = "vmess"
	Vless  = "vless"
	Trojan = "trojan"
	Socks  = "socks"
	Http   = "http"
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
			Settings: &Settings{
				Udp: true,
			},
		},
	}
	DemoOutbounds = []Bound{
		{
			Protocol: Vmess,
			Settings: &Settings{
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
	Listen         string          `json:"listen,omitempty"`
	Port           int             `json:"port,omitempty"`
	Protocol       string          `json:"protocol"`
	Settings       *Settings       `json:"settings"`
	StreamSettings *StreamSettings `json:"streamSettings,omitempty"`
}

type Settings struct {
	Udp     bool     `json:"udp,omitempty"`
	Vnext   []Vnext  `json:"vnext,omitempty"`
	Servers []Server `json:"servers,omitempty"`
}

type StreamSettings struct {
	Network         string          `json:"network"`
	Security        string          `json:"security"`
	RealitySettings RealitySettings `json:"realitySettings"`
}

type RealitySettings struct {
	ServerName    string `json:"serverName"`
	Fingerprint   string `json:"fingerprint"`
	Show          bool   `json:"show"`
	PublicKey     string `json:"publicKey"`
	ShortID       string `json:"shortId"`
	SpiderX       string `json:"spideX"`
	Mldsa64Verify string `json:"mldsa64Verify"`
}

type Server struct {
	Address  string `json:"address"`
	Port     int    `json:"port"`
	Password string `json:"password"`
}

type Vnext struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Users   []User `json:"users"`
}

type User struct {
	ID         string `json:"id"`
	AlterId    int    `json:"alterId"`
	Encryption string `json:"encryption"`
	Flow       string `json:"flow"`
}

/*



 */
