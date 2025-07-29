# Netpoller
> netpoller是如何做到同步非阻塞的？
> 通过golang的runtime scheduler加上操作系统的I/O多路复用，核心就是：让阻塞只阻塞在goroutine上，不要陷入内核。
Go 基于 I/O multiplexing 和 goroutine scheduler 构建了一个简洁而高性能的原生网络模型(基于 Go 的 I/O 多路复用 netpoller )，提供了 goroutine-per-connection 这样简单的网络编程模式。在这种模式下，开发者使用的是**同步的模式去编写异步的逻辑**
```go
func (srv *Server) Serve(l net.Listener) error {
	// .....
	for {
		rw, err := l.Accept()
		if err != nil {
			// .....
			return err
		}
		connCtx := ctx
		if cc := srv.ConnContext; cc != nil {
			connCtx = cc(connCtx, rw)
			if connCtx == nil {
				panic("ConnContext returned nil")
			}
		}
		tempDelay = 0
		c := srv.newConn(rw)
		c.setState(c.rwc, StateNew, runHooks) // before Serve can return
		go c.serve(connCtx) // 每个链接一个 goroutine
	}
}
```

## 参考
https://strikefreedom.top/archives/go-netpoll-io-multiplexing-reactor#%E5%AF%BC%E8%A8%80
https://xiaoxubeii.github.io/articles/linux-io-models-and-go-network-model/
https://lifan.tech/2020/07/03/golang/netpoll/