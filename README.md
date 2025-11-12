### 代理池
* 使用`xray`运行的代理池

### 如何使用
```go
    p := xray_pool.New(
		//xray_pool.WithSubscribe(s),
		xray_pool.WithNode(s),
	)
	defer p.Close()
    go p.Run()
	p.Do(func(n *xray_pool.Node) error {
		proxy:= n.Address()
        // do something
        return nil
    })
```