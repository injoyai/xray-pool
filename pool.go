package xray_pool

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/injoyai/base/types"
	"github.com/injoyai/logs"
)

const (
	DefaultConfigDir = "./config/"
	DefaultStartPort = 50000
	DefaultPoolCap   = 1000
	DefaultTimeout   = time.Second * 5
	ErrInvalidPort   = types.Err("端口被占用")
)

var (
	XrayCmd  = []string{"./bin/xray", "run", "-config"}
	V2rayCmd = []string{"v2ray", "-config"}
)

type Option func(*Pool)

func WithSubscribe(subscribe ...string) Option {
	return func(p *Pool) {
		p.subscribes = subscribe
	}
}

func WithNode(node ...string) Option {
	return func(p *Pool) {
		p.nodeUrls = append(p.nodeUrls, node...)
	}
}

func WithConfigDir(dir string) Option {
	return func(p *Pool) {
		p.configDir = dir
	}
}

func WithStartPort(port int) Option {
	return func(p *Pool) {
		p.startPort = port
	}
}

func WithNodeCheck(check CheckFunc) Option {
	return func(p *Pool) {
		p.nodeFunc = check
	}
}

func WithProxyCheck(check CheckFunc) Option {
	return func(p *Pool) {
		p.proxyCheck = check
	}
}

func WithPoolCap(cap int) Option {
	return func(p *Pool) {
		if cap >= 0 && cap != len(p.pool) {
			old := p.pool
			p.pool = make(chan *Node, cap)
			for n := range old {
				p.Put(n)
			}
			close(old)
		}
	}
}

// WithCmd 设置启动命令
// xray run -config
// v2ray -config
func WithCmd(cmd []string) Option {
	return func(p *Pool) {
		p.cmd = cmd
	}
}

func WithProtocol(protocol string) Option {
	return func(p *Pool) {
		p.protocol = protocol
	}
}

func New(op ...Option) *Pool {
	p := &Pool{
		configDir:  DefaultConfigDir,
		startPort:  DefaultStartPort,
		nodeFunc:   ByPing,
		proxyCheck: ByGoogle,
		pool:       make(chan *Node, DefaultPoolCap),
		cmd:        XrayCmd,
		protocol:   Mixed,
		done:       make(chan struct{}),
		started:    make(chan struct{}),
	}
	for _, o := range op {
		o(p)
	}
	return p
}

type Pool struct {
	subscribes []string  //订阅地址
	nodeUrls   []string  //节点地址
	configDir  string    //配置目录
	startPort  int       //起始端口
	nodeFunc   CheckFunc //检查节点是否可用,ping,tcp,download等
	proxyCheck CheckFunc //代理请求校验
	cmd        []string  //启动命令
	protocol   string    //协议

	pool     chan *Node        //代理池
	allNodes types.List[*Node] //全部节点
	valid    types.List[*Node] //有效节点
	done     chan struct{}     //
	once     sync.Once
	started  chan struct{}
}

func (p *Pool) Get() *Node {
	node := <-p.pool
	return node
}

func (p *Pool) Put(n *Node) {
	p.pool <- n
}

func (p *Pool) Len() int {
	return len(p.pool)
}

func (p *Pool) Started() <-chan struct{} {
	return p.started
}

func (p *Pool) Do(f func(n *Node) error) error {
	n := p.Get()
	defer p.Put(n)
	return f(n)
}

func (p *Pool) Run() error {
	p.Close()

	//获取所有节点地址,去重
	m := p.subscribe()

	//解析节点信息
	p.parse(m)

	//校验节点
	p.check()

	//开始启动
	p.start()

	p.done = make(chan struct{})
	p.once.Do(func() { close(p.done) })

	exitChan := make(chan os.Signal)
	signal.Notify(exitChan, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() { <-exitChan; p.Close() }()

	<-p.done

	return nil
}

func (p *Pool) Close() error {
	for _, n := range p.allNodes {
		n.Stop()
	}
	if p.done != nil {
		p.once.Do(func() { close(p.done) })
	}
	return nil
	//return os.RemoveAll(p.configDir)
}

func (p *Pool) subscribe() map[string]struct{} {
	m := make(map[string]struct{})
	for _, u := range p.subscribes {
		ls, err := p.get(u)
		if err != nil {
			logs.Err(err)
			continue
		}
		for _, l := range ls {
			if len(l) > 0 {
				m[l] = struct{}{}
			}
		}
	}
	for _, u := range p.nodeUrls {
		m[u] = struct{}{}
	}
	return m
}

func (p *Pool) get(u string) ([]string, error) {
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	ls := strings.Split(string(bs), "\n")
	return ls, nil
}

func (p *Pool) parse(m map[string]struct{}) {
	for u, _ := range m {
		n, err := p.parseNode(u)
		if err != nil {
			logs.Warn(err)
			continue
		}
		p.allNodes = append(p.allNodes, n)
	}
}

func (p *Pool) parseNode(u string) (*Node, error) {
	n := &Node{
		origin:     u,
		listenPort: -1,
		fail:       make(map[string]int),
	}
	if err := n.parse(); err != nil {
		return nil, err
	}
	return n, nil
}

func (p *Pool) check() {
	if p.nodeFunc == nil {
		p.nodeFunc = ByPing
	}
	wg := sync.WaitGroup{}
	for _, n := range p.allNodes {
		if n == nil {
			//logs.Debug("n is nil")
			continue
		}
		wg.Add(1)
		go func(n *Node) {
			defer wg.Done()
			err := n.check(p.nodeFunc)
			if err != nil {
				//logs.Warn(err)
				return
			}
			p.valid = append(p.valid, n)
		}(n)
	}
	wg.Wait()
	p.valid.Sort(func(a, b *Node) bool {
		return a.checkSpend < b.checkSpend
	})
}

func (p *Pool) start() {
	if len(p.valid) == 0 {
		return
	}
	os.MkdirAll(p.configDir, os.ModePerm)
	port := p.startPort
	wg := sync.WaitGroup{}
	wg.Add(len(p.valid))
	for _, n := range p.valid {
		go func(n *Node, port int) {
			defer wg.Done()
			if err := n.Start(port, p.protocol, p.configDir, p.cmd); err != nil {
				//logs.Warn(err)
				return
			}
			if p.proxyCheck != nil {
				if _, err := p.proxyCheck(n); err != nil {
					//logs.Warn(err)
					n.Stop()
					return
				}
			}
			logs.Info(n.Proxy(), "->", n.Origin())
			p.pool <- n
		}(n, port)
		port++
	}
	wg.Wait()
	close(p.started)
}

type Node struct {
	Vnexter

	origin         string         // 原始 vmess:// 链接
	listenProtocol string         // 本地 V2Ray 实例协议 http socks
	listenPort     int            // 本地 V2Ray 实例端口
	process        *exec.Cmd      // 本地 V2Ray 进程
	fail           map[string]int // 请求地址对应的失败次数
	failLimit      int            // 失败次数限制
	checkSpend     time.Duration  // 检查节点耗时

	running uint32
}

func (n *Node) String() string {
	return fmt.Sprintf("延迟:%s, %s", n.checkSpend, n.origin)
}

func (n *Node) Closed() bool {
	return atomic.LoadUint32(&n.running) == 0
}

func (n *Node) Origin() string {
	return n.origin
}

func (n *Node) Proxy() string {
	switch n.listenProtocol {
	case Socks:
		return fmt.Sprintf("socks5://127.0.0.1:%d", n.listenPort)
	case Http, Mixed:
		return fmt.Sprintf("http://127.0.0.1:%d", n.listenPort)
	}
	return fmt.Sprintf("socks5://127.0.0.1:%d", n.listenPort)
}

func (n *Node) parse() (err error) {
	switch {
	case strings.HasPrefix(n.origin, Vmess):
		n.Vnexter, err = ParseVmess(n.origin)
	case strings.HasPrefix(n.origin, Vless):
		n.Vnexter, err = ParseVless(n.origin)
	case strings.HasPrefix(n.origin, Trojan):
		n.Vnexter, err = ParseTrojan(n.origin)
	default:
		err = fmt.Errorf("invalid protocol, must be [vmess|vless|trojan]: %s", n.origin)
	}
	return
}

// Check 检查节点是否有效
func (n *Node) check(f CheckFunc) error {
	var err error
	n.checkSpend, err = f(n)
	return err
}

//func (n *Node) Fail(url string) {
//	n.fail[url] = n.fail[url] + 1
//	if n.failLimit > 0 && len(n.fail) > n.failLimit {
//		//todo
//	}
//	n.Stop()
//}

func (n *Node) Start(port int, protocol, configDir string, cmd []string) error {

	n.Stop()

	n.listenPort = port
	n.listenProtocol = protocol

	// 生成临时 V2Ray 配置
	config := Config{
		Log: DefaultLog,
		Inbounds: []Bound{
			{
				Port:     port,
				Listen:   "0.0.0.0",
				Protocol: protocol,
				Settings: &Settings{
					Udp: true,
				},
			},
		},
		Outbounds: []Bound{
			{
				Protocol:       n.Protocol(),
				Settings:       n.Settings(),
				StreamSettings: n.StreamSettings(),
			},
		},
	}

	// 保存临时配置
	file := fmt.Sprintf(configDir+"%d.json", port)
	if err := os.WriteFile(file, config.Bytes(), 0644); err != nil {
		return err
	}

	// 启动 V2Ray/Xray
	name, args := cmd[0], append(cmd[1:], file)
	c := exec.Command(name, args...)

	stdout, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	defer stdout.Close()

	go func() {
		err := c.Run()
		_ = err
		//logs.PrintErr(err)
		atomic.StoreUint32(&n.running, 0)
	}()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		//等待成功启动
		switch {
		case strings.Contains(line, "[Info] infra/conf/serial: Reading config:"):
			atomic.StoreUint32(&n.running, 1)
			n.process = c
			return nil
		case strings.Contains(line, "bind: Only one usage of each socket address"):
			return ErrInvalidPort
		}
	}

	return nil
}

func (n *Node) Stop() error {
	if n.process == nil || n.process.Process == nil {
		return nil
	}
	err := n.process.Process.Kill()
	if err != nil {
		return err
	}
	n.listenPort = -1
	n.checkSpend = -1
	n.process = nil
	return nil
}
