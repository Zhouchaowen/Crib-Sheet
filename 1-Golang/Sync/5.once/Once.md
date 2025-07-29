# Once

> 通过Mutex+原子类实现,atomic判断是否调用过，没有调用则加锁，再次检测atomic后调用f，调用后把atomic置为true

```go
type Once struct {
	_ noCopy

	done atomic.Bool
	m    Mutex
}

func (o *Once) Do(f func()) {
	
	if !o.done.Load() {
		// Outlined slow-path to allow inlining of the fast-path.
		o.doSlow(f)
	}
}

func (o *Once) doSlow(f func()) {
	o.m.Lock()
	defer o.m.Unlock()
	if !o.done.Load() {
		defer o.done.Store(true)
		f()
	}
}

```