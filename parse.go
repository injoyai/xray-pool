package xray_pool

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/injoyai/conv"
)

type Vnexter interface {
	Remark() string
	Protocol() string
	Hostname() string
	Port() int
	Settings() *Settings
	StreamSettings() *StreamSettings
}

/*
ParseVless
vless://c37cdcff-42f1-4f09-8ce1-0df6cf7e2520@sandking.fonixapp.org:33115?encryption=none&flow=xtls-rprx-vision&security=reality&sni=yelp.com&fp=chrome&pbk=53Q1y0Vmf2zaGBBlcO1NyKFvQM1TShkJKBCNjlevpns&sid=09cb&spx=%2F&allowInsecure=1&type=tcp&headerType=none#%F0%9F%87%A6%F0%9F%87%B9%20www.85.com%20%E5%A5%A5%E5%9C%B0%E5%88%A9
*/
func ParseVless(raw string) (*VlessConfig, error) {
	raw = strings.TrimPrefix(raw, "vless://")
	// 分离备注
	parts := strings.SplitN(raw, "#", 2)
	link := parts[0]
	remark := ""
	if len(parts) > 1 {
		remark, _ = url.QueryUnescape(parts[1])
	}

	u, err := url.Parse("vless://" + link)
	if err != nil {
		return nil, err
	}

	return &VlessConfig{
		remark:   remark,
		hostname: u.Hostname(),
		port:     conv.Int(u.Port()),
		settings: &Settings{
			Vnext: []Vnext{
				{
					Address: u.Hostname(),
					Port:    conv.Int(u.Port()),
					Users: []User{
						{
							ID:         u.User.Username(),
							AlterId:    0,
							Encryption: u.Query().Get("encryption"),
							Flow:       u.Query().Get("flow"),
						},
					},
				},
			},
		},
		streamSettings: &StreamSettings{
			Network:  u.Query().Get("type"),
			Security: u.Query().Get("security"),
			RealitySettings: RealitySettings{
				ServerName:    u.Query().Get("sni"),
				Fingerprint:   u.Query().Get("fp"),
				Show:          false,
				PublicKey:     u.Query().Get("pbk"),
				ShortID:       u.Query().Get("sid"),
				SpiderX:       u.Query().Get("spx"),
				Mldsa64Verify: "",
			},
		},
	}, nil
}

type VlessConfig struct {
	remark         string
	hostname       string
	port           int
	settings       *Settings
	streamSettings *StreamSettings
}

func (c *VlessConfig) Remark() string {
	return c.remark
}

func (c *VlessConfig) Protocol() string {
	return Vless
}

func (c *VlessConfig) Hostname() string {
	return c.hostname
}

func (c *VlessConfig) Port() int { return c.port }

func (c *VlessConfig) Settings() *Settings {
	return c.settings
}

func (c *VlessConfig) StreamSettings() *StreamSettings {
	return c.streamSettings
}

/*



 */

/*
ParseVmess
vmess://eyJwcyI6Ind3dy44NWxhLmNvbfCfh7rwn4e4VVNfNTR8ODc5S0IvcyIsImFkZCI6InNzc3Nzc3Nzc3Nzc2ZmZmZmZmZnaC4yMDMyLnBwLnVhIiwiYWlkIjowLCJpZCI6IjQxNzRiOTVkLTExNWUtNGQzOS1hZGQ2LTFmOGRiOTViYjg2MCIsIm5ldCI6IndzIiwic2N5IjoiYXV0byIsInBvcnQiOjQ0MywidGxzIjoidGxzIiwicGF0aCI6Ii82V2UzVTlEZjFXR3hnRm5vRlB3MSIsImhvc3QiOiJzc3Nzc3Nzc3Nzc3NmZmZmZmZmZ2guMjAzMi5wcC51YSIsInNuaSI6InNzc3Nzc3Nzc3Nzc2ZmZmZmZmZnaC4yMDMyLnBwLnVhIn0=
{"ps":"www.85la.comUS_54|879KB/s","add":"ssssssssssssfffffffgh.2032.pp.ua","aid":0,"id":"4174b95d-115e-4d39-add6-1f8db95bb860","net":"ws","scy":"auto","port":443,"tls":"tls","path":"/6We3U9Df1WGxgFnoFPw1","host":"ssssssssssssfffffffgh.2032.pp.ua","sni":"ssssssssssssfffffffgh.2032.pp.ua"}
*/
func ParseVmess(u string) (*VmessConfig, error) {
	raw := strings.TrimPrefix(u, "vmess://")
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	var cfg VmessConfig
	err = json.Unmarshal(data, &cfg)
	return &cfg, err
}

type VmessConfig struct {
	Hostname_ string `json:"host"`
	Port_     string `json:"port"`
	UID       string `json:"id"`
	AlterID   string `json:"aid"`
	Remark_   string `json:"ps"`
	Network   string `json:"net"`
	Path      string `json:"path"`
	Security  string `json:"scy"`
	SNI       string `json:"sni"`
	TLS       string `json:"tls"`
}

func (c *VmessConfig) Remark() string {
	return c.Remark_
}

func (c *VmessConfig) Protocol() string {
	return Vmess
}

func (c *VmessConfig) Hostname() string {
	return c.Hostname_
}

func (c *VmessConfig) Port() int { return conv.Int(c.Port_) }

func (c *VmessConfig) Settings() *Settings {
	return &Settings{
		Vnext: []Vnext{
			{
				Address: c.Hostname_,
				Port:    c.Port(),
				Users: []User{{
					ID:         c.UID,
					AlterId:    conv.Int(c.AlterID),
					Encryption: None,
				}},
			},
		},
	}
}

func (c *VmessConfig) StreamSettings() *StreamSettings {
	return &StreamSettings{
		Network:  c.Network,
		Security: c.Security,
		RealitySettings: RealitySettings{
			ServerName: c.SNI,
		},
	}
}

/*



 */

// ParseTrojan 解析 trojan:// 链接
// trojan://slch2024@190.93.244.87:2096?type=ws&sni=ocost-dy.wmlefl.cc&allowInsecure=1&path=/Telegram🇨🇳&host=ocost-dy.wmlefl.cc#www.85la.com%F0%9F%87%BA%F0%9F%87%B8US_52%7C1.4MB%2Fs
func ParseTrojan(raw string) (*TrojanConfig, error) {
	raw = strings.TrimPrefix(raw, "trojan://")

	// 拆分备注部分
	parts := strings.SplitN(raw, "#", 2)
	link := parts[0]
	remark := conv.Default("", parts[1:]...)

	// 为了能用 url.Parse 解析 user@host:port?query 的结构，补上协议头
	u, err := url.Parse("trojan://" + link)
	if err != nil {
		return nil, err
	}

	// password 可能在 User.Username() 中（常见形式），也可能在 query 中（不常见）
	password := conv.Default(u.User.Username(), u.Query()["password"]...)

	// 端口默认为空则用 443
	port := conv.Select(u.Port() == "", 443, conv.Int(u.Port()))

	// 常用参数：sni / alpn / plugin 等
	//sni := conv.Default(u.Hostname(), u.Query()["sni"]...)
	//alpn := u.Query().Get("alpn")
	//plugin := u.Query().Get("plugin")

	return &TrojanConfig{
		remark:   remark,
		hostname: u.Hostname(),
		port:     conv.Int(port),
		settings: &Settings{
			Servers: []Server{
				{
					Address:  u.Hostname(),
					Port:     conv.Int(port),
					Password: password,
				},
			},
		},
	}, nil
}

type TrojanConfig struct {
	remark         string
	hostname       string
	port           int
	settings       *Settings
	streamSettings *StreamSettings
}

func (c *TrojanConfig) Remark() string   { return c.remark }
func (c *TrojanConfig) Protocol() string { return Trojan }

func (c *TrojanConfig) Hostname() string {
	return c.settings.Servers[0].Address
}

func (c *TrojanConfig) Port() int { return c.port }

func (c *TrojanConfig) Settings() *Settings {
	return c.settings
}

func (c *TrojanConfig) StreamSettings() *StreamSettings {
	return c.streamSettings
}
