## channel是如何实现在不同的 goroutine 之间传递数据的？
[参考](https://lxb.wiki/7b2461e3/)

```
type hchan struct {
    qcount   uint           // 队列中的元素数量
    dataqsiz uint           // 环形队列大小（缓冲区大小）
    buf      unsafe.Pointer // 缓冲区的指针
    elemsize uint16         // 元素大小
    closed   uint32         // 通道是否关闭
    timer    *timer         // 用于定时器的通道
    elemtype *_type         // 元素类型
    sendx    uint           // 发送索引
    recvx    uint           // 接收索引
    recvq    waitq          // 接收等待队列
    sendq    waitq          // 发送等待队列
    ...
    lock mutex              // 互斥锁，保护上述字段
}
```
    channle是如何知道goroutine在接收和发送数据的？
    接收队列的goroutine是如何添加进去的？

## channel是如何实现安全地进行并发操作？
## channel是如何实现阻塞性的？
## 为什么创建过多的 channel，可能会影响性能？

## channel本质是基于锁的先进先出队列？