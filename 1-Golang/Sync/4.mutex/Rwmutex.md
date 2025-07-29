# Rwmutex
> 通过mutex+原子操作和信号量机制，实现高效的多读单写互斥。

```go
type RWMutex struct {
    w           Mutex        // 写锁，互斥于其他写者
    writerSem   uint32       // 写操作信号量
    readerSem   uint32       // 读操作信号量
    readerCount atomic.Int32 // 活跃读者计数（正数=无写者，负数=有写者）
    readerWait  atomic.Int32 // 离开中的读者数
}
```
## 读锁
```go
func (rw *RWMutex) RLock() {
	//....

    // 活跃读者的数量。>0 代表有读者，<0 代表有写者（或写者在等待）。
	if rw.readerCount.Add(1) < 0 {
		// A writer is pending, wait for it.
        // <0 当前读者阻塞等待
		runtime_SemacquireRWMutexR(&rw.readerSem, false, 0)
	}

	//....
}

func (rw *RWMutex) RUnlock() {
	//....

    // 如果减一后的结果是正数，说明没有写者在等待（即没有调用 Lock()），直接返回即可，这就是快路径。
	if r := rw.readerCount.Add(-1); r < 0 {
		// Outlined slow-path to allow the fast-path to be inlined
        // < 0 说明此时有写者在等待，进入 rUnlockSlow
        // rUnlockSlow的作用是判定是否是最后一个读者，是最后一个读者的话，唤醒写者
		rw.rUnlockSlow(r)
	}

	//....
}

func (rw *RWMutex) rUnlockSlow(r int32) {
	if r+1 == 0 || r+1 == -rwmutexMaxReaders {
		race.Enable()
		fatal("sync: RUnlock of unlocked RWMutex")
	}

    // 每有一个读者释放锁，就 readerWait 减一。
    // 最后一个读者释放时（readerWait 变为 0），说明写者可以独占锁了，于是调用 runtime_Semrelease 唤醒等待的写者。
	// A writer is pending.
	if rw.readerWait.Add(-1) == 0 {
		// The last reader unblocks the writer.
		runtime_Semrelease(&rw.writerSem, false, 1)
	}
}
```


## 写锁
```go
func (rw *RWMutex) Lock() {
	//....
	// First, resolve competition with other writers.
	rw.w.Lock()

	// Announce to readers there is a pending writer.
    // 一下子减去一个极大的数（变成负数），标记此时有写者要来/已来，阻止新的读者进入;
    // 后续的 RLock 会发现 readerCount 是负数，进入等待。
	r := rw.readerCount.Add(-rwmutexMaxReaders) + rwmutexMaxReaders
	// Wait for active readers.
	if r != 0 && rw.readerWait.Add(r) != 0 {
        // 如果还有活跃读者，写者阻塞自己, 然后等待最后一个读者唤醒自己（rUnlockSlow）
		runtime_SemacquireRWMutex(&rw.writerSem, false, 0)
	}
	//....
}

func (rw *RWMutex) Unlock() {
	//....

	// Announce to readers there is no active writer.
	r := rw.readerCount.Add(rwmutexMaxReaders)
    // 如果此时 r >= rwmutexMaxReaders，说明当前 RWMutex 根本没有持有写锁，却调用了解锁操作，这是不合法的！
	if r >= rwmutexMaxReaders {
        // 为什么会出现这种情况？
        // 只有在没有写锁的情况下，readerCount 才可能大于等于 0。
        // 此时再加上 rwmutexMaxReaders，就变成了一个非常大的数（大于等于 rwmutexMaxReaders）。
        // 也就是说，如果你多次 Unlock，或者从未 Lock 过就 Unlock，就会触发这个判断并 panic。
		race.Enable()
		fatal("sync: Unlock of unlocked RWMutex")
	}
	// Unblock blocked readers, if any.
    // 通知读者没有写者，恢复 readerCount
	for i := 0; i < int(r); i++ {
		runtime_Semrelease(&rw.readerSem, false, 0)
	}
	// Allow other writers to proceed.
	rw.w.Unlock()
	//....
}
```