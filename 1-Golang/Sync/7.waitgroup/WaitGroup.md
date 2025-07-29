# WaitGroup
> 用 state 原子变量将 counter（高32位）和 waiter（低32位）合并存储，利用原子操作保持并发安全。
> 用信号量 sema 负责 goroutine 的阻塞和唤醒。
WaitGroup 用于“等待一组 goroutine 结束”。常用于并发执行任务后，等待所有任务完成后再进行下一步操作。
WaitGroup 不能复制，必须用指针传递。
Add(1) 必须在启动新 goroutine 之前，且 Wait 必须在所有 Add(1) 之后调用。
Add、Done、Wait 都用了原子操作保证并发安全。
如果 Add(1) 和 Wait() 并发调用（计数为0时），会 panic。
```go
type WaitGroup struct {
	noCopy noCopy

	// Bits (high to low):
	//   bits[0:32]  counter
	//   bits[32]    flag: synctest bubble membership
	//   bits[33:64] wait count
	state atomic.Uint64
	sema  uint32 // 信号量用于唤醒等待的 goroutine
}
```

## Add
调用Add就会在state的高位上进行计数（增加或者减少），当减少为0后就会获取低位上正在等待的goroutine的数量，然后循环调用runtime_Semrelease来唤醒被阻塞的goroutine

## Wait
调用Wait后，会先读取state的低位如果等于0就直接返回，如果不为0，则+1后通过runtime_SemacquireWaitGroup阻塞，等待被ADD操作后唤醒。
