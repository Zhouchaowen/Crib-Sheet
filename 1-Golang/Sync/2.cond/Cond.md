# Cond
> Cond 适合用于复杂的同步/等待场景，替代简单场景更推荐用 channel。如果你遇到“等待条件成立再继续执行”的需求，Cond 就非常适合。
Cond 必须搭配锁（Mutex/RWMutex）使用。
Wait 前后要加锁，Wait 内部自动释放锁，并在被唤醒后重新加锁。
不能复制 Cond 对象，否则会 panic。
多个 goroutine 等待时，用 Signal 唤醒一个，用 Broadcast 唤醒所有。


## 介绍一下Golang核心特性之一：sync.Cond?

