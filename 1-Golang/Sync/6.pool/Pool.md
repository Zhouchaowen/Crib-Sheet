# Pool
> 每个 P（即调度器 Processor，GOMAXPROCS 决定其数量）分配一个本地 poolLocal 结构体。这样可以极大减少锁争用，因为同一个 P 内的 goroutine 访问和操作本地 pool 时无需加锁。
> 通过“每 P 本地化+work stealing+GC 清理”机制，实现了高效、低锁争用、内存可控的对象池，非常适合临时复用对象，避免频繁分配和回收带来的性能损耗。
```go
type Pool struct {
	noCopy noCopy

	local     unsafe.Pointer // local fixed-size per-P pool, actual type is [P]poolLocal
	localSize uintptr        // size of the local array

	victim     unsafe.Pointer // local from previous cycle
	victimSize uintptr        // size of victims array

	// New optionally specifies a function to generate
	// a value when Get would otherwise return nil.
	// It may not be changed concurrently with calls to Get.
	New func() any
}

type poolLocalInternal struct {
	private any       // Can be used only by the respective P.
	shared  poolChain // Local P can pushHead/popHead; any P can popTail.
}

type poolLocal struct {
	poolLocalInternal

	// Prevents false sharing on widespread platforms with
	// 128 mod (cache line size) = 0 .
	pad [128 - unsafe.Sizeof(poolLocalInternal{})%128]byte
}
```

## Put 操作过程
优先把对象放到当前 P 的 private，如果 private 已有对象，则放入 shared 队列头部。
在 race 检测下有一定概率直接丢弃对象，以防止池无限增长。
```go
func (p *Pool) Put(x any) {
	if x == nil {
		return
	}
	// race....
	l, _ := p.pin()
	if l.private == nil {
		l.private = x
	} else {
		l.shared.pushHead(x)
	}
	runtime_procUnpin()
	// race....
}
```

## Get 操作过程
先从当前 P 的 private 获取对象，若无则从 shared 队列头部获取。
如果本地都没有，则尝试从其它 P 的 shared 尾部“偷”对象（work stealing）。
实在没有，则调用 Pool.New（如果设置了）构造新对象，否则返回 nil。
```go
func (p *Pool) Get() any {
	// race....
	l, pid := p.pin()
	x := l.private
	l.private = nil
	if x == nil {
		// Try to pop the head of the local shard. We prefer
		// the head over the tail for temporal locality of
		// reuse.
		x, _ = l.shared.popHead()
		if x == nil {
			x = p.getSlow(pid)
		}
	}
	runtime_procUnpin()
	// race....
	if x == nil && p.New != nil {
		x = p.New()
	}
	return x
}
```