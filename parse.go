package xray_pool

import (
	"encoding/base64"
	"encoding/json"
	"github.com/injoyai/conv"
	"net/url"
	"strings"
)

type Vnexter interface {
	Remark() string
	Protocol() string
	Hostname() string
	Port() int
	Vnext() Vnext
}

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
		UUID:      u.User.Username(),
		Hostname_: u.Hostname(),
		Port_:     u.Port(),
		Security:  u.Query().Get("security"),
		Type:      u.Query().Get("type"),
		SNI:       u.Query().Get("sni"),
		Path:      u.Query().Get("path"),
		Remark_:   remark,
	}, nil
}

type VlessConfig struct {
	UUID      string
	Hostname_ string
	Port_     string
	Security  string
	Type      string
	SNI       string
	Path      string
	Remark_   string
}

func (c *VlessConfig) Remark() string {
	return c.Remark_
}

func (c *VlessConfig) Protocol() string {
	return Vless
}

func (c *VlessConfig) Hostname() string {
	return c.Hostname_
}

func (c *VlessConfig) Port() int {
	return conv.Int(c.Port_)
}

func (c *VlessConfig) Vnext() Vnext {
	return Vnext{
		Address: c.Hostname(),
		Port:    c.Port(),
		Users: []User{{
			ID: c.UUID,
		}},
	}
}

/*



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
	Hostname_ string `json:"add"`
	Port_     string `json:"port"`
	ID        string `json:"id"`
	AlterId   string `json:"aid"`
	Remark_   string `json:"ps"`
	Network   string `json:"net"`
	Path      string `json:"path"`
	Security  string `json:"scy"`
	Type      string `json:"type"`
	SNI       string `json:"sni"`
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

func (c *VmessConfig) Port() int {
	return conv.Int(c.Port_)
}

func (c *VmessConfig) Vnext() Vnext {
	return Vnext{
		Address: c.Hostname(),
		Port:    c.Port(),
		Users: []User{{
			ID:      c.ID,
			AlterId: conv.Int(c.AlterId),
		}},
	}
}

/*



 */

// ParseTrojan 解析 trojan:// 链接
// 形式通常为:
// trojan://password@host:port?param1=val1&...#remark
func ParseTrojan(raw string) (*TrojanConfig, error) {
	raw = strings.TrimPrefix(raw, "trojan://")

	// 拆分备注部分
	parts := strings.SplitN(raw, "#", 2)
	link := parts[0]
	remark := ""
	if len(parts) > 1 {
		remark, _ = url.QueryUnescape(parts[1])
	}

	// 为了能用 url.Parse 解析 user@host:port?query 的结构，补上协议头
	u, err := url.Parse("trojan://" + link)
	if err != nil {
		return nil, err
	}

	// password 可能在 User.Username() 中（常见形式），也可能在 query 中（不常见）
	password := u.User.Username()
	if password == "" {
		// 有些实现会把 password 放到 query 参数 password 中
		password = u.Query().Get("password")
	}

	// 端口默认为空则用 443
	port := u.Port()
	if port == "" {
		port = "443"
	}

	// 常用参数：sni / alpn / plugin 等
	sni := u.Query().Get("sni")
	if sni == "" {
		sni = u.Hostname()
	}
	alpn := u.Query().Get("alpn")
	plugin := u.Query().Get("plugin")

	return &TrojanConfig{
		Password:  password,
		Hostname_: u.Hostname(),
		Port_:     port,
		SNI:       sni,
		ALPN:      alpn,
		Plugin:    plugin,
		Remark_:   remark,
	}, nil
}

type TrojanConfig struct {
	Password  string
	Hostname_ string
	Port_     string
	SNI       string
	ALPN      string
	Plugin    string
	Remark_   string
}

func (c *TrojanConfig) Remark() string   { return c.Remark_ }
func (c *TrojanConfig) Protocol() string { return Trojan } // 假定你在常量中已定义 Trojan = "trojan"
func (c *TrojanConfig) Hostname() string { return c.Hostname_ }
func (c *TrojanConfig) Port() int        { return conv.Int(c.Port_) }

func (c *TrojanConfig) Vnext() Vnext {
	return Vnext{
		Address: c.Hostname(),
		Port:    c.Port(),
		Users: []User{{
			ID: c.Password, // 把 trojan 的 password 放到 User.ID 字段以兼容上层
		}},
	}
}
