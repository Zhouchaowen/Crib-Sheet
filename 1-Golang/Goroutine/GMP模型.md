## 什么是进程
> 进程本质上是操作系统管理的一个资源集合，又因为linux上一切皆文件，所以这个资源集合其实就是 /proc/<PID> 目录下的这些文件

更准确的理解,进程是操作系统管理的资源集合，这些资源实际存在于：
1. 物理内存
2. CPU寄存器
3. 内核数据结构
4. 其他硬件资源

/proc/<PID> 目录是这些资源的一个"视图"：
1. 提供了查看和操作这些资源的接口
2. 是内核将进程信息暴露给用户空间的方式
3. 不是资源本身，而是资源的映射

"一切皆文件"是Linux的设计哲学：
1. 提供统一的访问接口
2. 简化了资源的管理和操作
3. 但实际资源并不是真正的文件

这就像看房子：
1. 房子本身是实际存在的（进程的实际资源）
2. 房产证和各种文档是房子的记录（/proc文件系统）
3. 文档帮助我们了解和管理房子，但不是房子本身

## 什么是线程
线程本质上也是一个特殊的进程（轻量级进程，LWP：Light Weight Process），它与其他同进程的线程共享某些资源。

通过这个程序可以观察操作系统下的线程
```
package main

import (
    "fmt"
    "os"
    "runtime"
    "sync"
    "time"
)

func main() {
    fmt.Printf("主进程 PID: %d\n", os.Getpid())
    
    var wg sync.WaitGroup
    
    // 创建3个线程
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            // 获取当前goroutine的系统线程ID
            fmt.Printf("线程 %d 运行中...\n", id)
            time.Sleep(time.Hour) // 保持线程运行
        }(i)
    }

    fmt.Println("\n现在可以通过以下命令观察线程：")
    fmt.Printf("1. ps -T -p %d\n", os.Getpid())
    fmt.Printf("2. ls -l /proc/%d/task\n", os.Getpid())
    
    wg.Wait()
}
```
所以，线程可以理解为：
1. 是一个特殊的轻量级进程
2. 与同进程的其他线程共享大部分资源
3. 在Linux中通过 task_struct 管理
4. 可以通过 /proc 文件系统观察
5. 拥有自己的执行上下文和栈空间

## 为什么会有Goroutine

> 传统模型的问题

**1:1 模型的局限**
```
用户线程 1 -----> 内核线程 1
用户线程 2 -----> 内核线程 2
用户线程 3 -----> 内核线程 3
```
1. 创建成本高
2. 切换开销大

**N:1 模型的局限**
```
用户线程 1 -\
用户线程 2 ---> 内核线程 1
用户线程 3 -/
```
1. 无法利用多核
2. 一个线程阻塞会影响所有线程

**M:N 模型的优势**
1. 灵活的调度
2. 更好的利用多核

## 阻塞时一切罪恶的源泉
阻塞本质上是等待某个条件满足，主要有以下几种情况：
```
// 1. IO阻塞
func ioBlock() {
    file, _ := os.Open("large_file.txt")
    data := make([]byte, 1024)
    file.Read(data)  // 等待磁盘IO
}

// 2. 网络阻塞
func networkBlock() {
    resp, _ := http.Get("http://example.com")  // 等待网络响应
}

// 3. 锁阻塞
func lockBlock() {
    var mu sync.Mutex
    mu.Lock()
    // ... 其他地方持有锁
    mu.Lock()  // 等待锁释放
}

// 4. 睡眠阻塞
func sleepBlock() {
    time.Sleep(time.Second)  // 主动睡眠
}
```

实际例子
```
func main() {
    // 创建两个goroutine
    go func() {
        // Goroutine 1: IO操作
        file, _ := os.Open("large_file.txt")
        data := make([]byte, 1024)
        file.Read(data)  // 阻塞点1
    }()
    
    go func() {
        // Goroutine 2: 计算操作
        sum := 0
        for i := 0; i < 1000000; i++ {
            sum += i
        }
    }()
}
```
当发生阻塞时的处理流程：
```
1. IO阻塞发生：
   Goroutine 1 执行 file.Read()
   |
   |--> 系统调用 sys_read()
        |
        |--> 数据未就绪
             |
             |--> 线程进入睡眠状态
                  |
                  |--> 调度器介入
                       |
                       |--> 切换到其他线程
```

## 理解GMP
[GMP 并发调度器深度解析之手撸一个高性能 goroutine pool](https://strikefreedom.top/archives/high-performance-implementation-of-goroutine-pool)


## 最佳实际

1.P的数量设置
```go
// CPU 密集型任务：设置为物理核心数
runtime.GOMAXPROCS(runtime.NumCPU())

// IO 密集型任务：可以设置更大的值
runtime.GOMAXPROCS(runtime.NumCPU() * 2)
```

2.控制Goroutine的数量
```go
// 使用 channel 控制并发数
semaphore := make(chan struct{}, maxConcurrency)

for i := 0; i < tasks; i++ {
    semaphore <- struct{}{} // 获取令牌
    go func() {
        defer func() { <-semaphore }() // 释放令牌
        // 任务处理
    }()
}
```

3.使用工作池模式
```go
// 不好的示例：无限制创建 goroutine
func badExample() {
    for i := 0; i < 100000; i++ {
        go func() {
            // 一些操作
        }()
    }
}

// 好的示例：使用工作池模式
type Pool struct {
    queue    chan func()
    workers  int
    wg       sync.WaitGroup
}

func NewPool(workers int) *Pool {
    p := &Pool{
        queue:   make(chan func(), workers),
        workers: workers,
    }
    p.start()
    return p
}

func (p *Pool) start() {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go func() {
            defer p.wg.Done()
            for task := range p.queue {
                task()
            }
        }()
    }
}

func (p *Pool) Submit(task func()) {
    p.queue <- task
}

func (p *Pool) Wait() {
    close(p.queue)
    p.wg.Wait()
}
```