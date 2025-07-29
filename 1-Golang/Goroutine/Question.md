# 疑惑
Goroutine的大致创建过程
https://www.cnblogs.com/flhs/p/12677335.html
https://cbsheng.github.io/posts/%E6%8E%A2%E7%B4%A2goroutine%E7%9A%84%E5%88%9B%E5%BB%BA/
https://golang.design/go-questions/sched/main-goroutine/
##　分配给一个M（线程）执行，执行的具体是什么东西？编译后的机器码？
    源代码 -> 词法分析 -> 语法分析 -> AST -> SSA -> 机器码
    编译后的机器码存储在可执行文件的代码段（text section）中
    每个函数都有其对应的入口地址
    G结构体中会保存这个入口地址（程序计数器PC值）
```
   M（系统线程）
     |
     |--> 获取G的上下文
     |    - 程序计数器（PC）
     |    - 栈指针（SP）
     |    - 其他寄存器状态
     |
     |--> 加载机器码
     |    - 从代码段加载对应位置的机器码
     |
     |--> 执行机器码
          - CPU直接执行机器指令
```
当需要切换Goroutine时：
```
     保存当前G的状态
       |
       |--> 保存PC（程序计数器）
       |--> 保存SP（栈指针）
       |--> 保存寄存器
       
     加载新G的状态
       |
       |--> 恢复PC
       |--> 恢复SP
       |--> 恢复寄存器
```

## 发生阻塞时，调度器时如何发现阻塞的，阻塞完成后又是如何切换回来的？
系统调用阻塞
```
   G1准备进行系统调用
     |
     |--> runtime检测到系统调用
          |
          |--> P解绑定当前M1（与M1分离）
          |    |--> P寻找空闲或创建新的M2
          |    |--> P与M2绑定继续执行其他G
          |
          |--> M1继续执行G1的系统调用（阻塞中）
               |
               |--> 系统调用完成
                    |--> M1尝试获取P（可能是原来的P，也可能是其他P）
                         |--> 获取成功：继续执行G1
                         |--> 获取失败：G1放入全局队列，M1放回空闲队列
```
    - 执行过程中会导致P与M解绑？P会寻找其他M或创建M，这是否会导致创建许多M？M可以背销毁吗？什么时候销毁的？
    - 执行过程中会导致P与M解绑，解绑后这个G还在M上执行？系统调用完成后的结果给谁了？
```
   系统调用结果的传递：
   1. 结果保存在G1的栈上
   2. G1要么：
      - 在获取到P后继续执行
      - 或被放入全局队列等待其他P执行
```
```
   初始状态：
   P1 --- M1 --- G1
   
   系统调用开始：
   P1 --- M2 --- G2    M1 --- G1(syscall)
   
   系统调用完成（成功获取P的情况）：
   P2 --- M1 --- G1    P1 --- M2 --- G2
   
   系统调用完成（未获取P的情况）：
   P1 --- M2 --- G2    
   Global Queue <- G1
   Idle List <- M1
```
    - 系统调用完成,M1尝试获取P,M1怎么知道系统调用完成了？M1是通过什么方式来获取P的？
    - 阻塞检测机制是什么？
        特定操作点都有内置的检测机制

同步阻塞
```
   G1准备进行channel接收
     |
     |--> runtime检测到channel阻塞
          |
          |--> G1状态改为waiting
          |--> G1从P的运行队列移除
          |--> P继续执行其他G
          |
          |--> channel准备就绪
               |--> G1状态改为runnable
               |--> G1重新加入P的运行队列或全局队列
```
    - G1怎么知道channel准备就绪了？
        - channel是如何唤醒G1的
            唤醒过程是由channel操作触发的
            调度器负责将被唤醒的G放入运行队列
            G的上下文会被完整保存和恢复
            整个过程是原子的（通过锁保护）
```
   // 简化的唤醒流程
   G1(阻塞) --> Channel操作 --> 调度器唤醒 --> G1(就绪) --> G1(运行)
```
``` 阻塞发送过程：
   // 当G1尝试向channel发送数据
   func chansend(c *hchan, elem unsafe.Pointer, block bool) bool {
       // 1. 获取channel锁
       lock(&c.lock)
       
       if c.closed != 0 {
           unlock(&c.lock)
           panic("send on closed channel")
       }
       
       // 2. 检查是否有等待接收的G
       if sg := c.recvq.dequeue(); sg != nil {
           // 直接发送给等待的接收者
           send(c, sg, elem)
           unlock(&c.lock)
           return true
       }
       
       // 3. 如果channel有缓冲区且未满
       if c.qcount < c.dataqsiz {
           // 将数据复制到缓冲区
           // ...
           unlock(&c.lock)
           return true
       }
       
       // 4. 无法立即发送，将当前G放入等待队列
       gp := getg()
       mysg := acquireSudog()
       mysg.elem = elem
       mysg.g = gp
       c.sendq.enqueue(mysg)
       
       // 5. 当前G进入休眠
       gopark(chanparkcommit, unsafe.Pointer(&c.lock), waitReasonChanSend, traceEvGoBlockSend, 2)
       
       // 6. 被唤醒后继续执行
   }
```
``` 阻塞接收过程：
   // 当G2尝试从channel接收数据
   func chanrecv(c *hchan, elem unsafe.Pointer, block bool) bool {
       // 1. 获取channel锁
       lock(&c.lock)
       
       // 2. 检查是否有等待发送的G
       if sg := c.sendq.dequeue(); sg != nil {
           // 直接从发送者接收数据
           recv(c, sg, elem)
           unlock(&c.lock)
           return true
       }
       
       // 3. 如果channel有数据
       if c.qcount > 0 {
           // 从缓冲区读取数据
           // ...
           unlock(&c.lock)
           return true
       }
       
       // 4. 无法立即接收，将当前G放入等待队列
       gp := getg()
       mysg := acquireSudog()
       mysg.elem = elem
       mysg.g = gp
       c.recvq.enqueue(mysg)
       
       // 5. 当前G进入休眠
       gopark(chanparkcommit, unsafe.Pointer(&c.lock), waitReasonChanReceive, traceEvGoBlockRecv, 2)
       
       // 6. 被唤醒后继续执行
   }
```

## 为什么上下文切换快（不需要切换到内核态）
```
   线程切换涉及特权级别变化：
   用户态(Ring 3) -> 系统调用 -> 内核态(Ring 0) -> 上下文切换 -> 用户态(Ring 3)
   
   Goroutine切换保持在同一特权级别：
   用户态(Ring 3) -> 切换Goroutine -> 用户态(Ring 3)
```
## GMP和CPU的关系，是如何充分利用多核CPU？
```
   G (Goroutine): 用户级线程，代表一个并发任务
   P (Processor): 处理器，代表执行Go代码所需的资源
   M (Machine): 系统线程，可以理解为内核线程
   CPU: 物理处理器核心
   
   工作关系：
   G --- 运行在 --> P --- 绑定在 --> M --- 运行在 --> CPU核心
```
```
   // P的数量
   func init() {
       // 默认P的数量等于CPU核心数
       GOMAXPROCS = runtime.NumCPU()
   }
   
   // 实际运行时
   P的数量 = GOMAXPROCS    // 可配置，默认等于CPU核心数
   M的数量 <= P的数量 + 阻塞M数量  // 动态调整
   G的数量 >> P的数量      // 远大于P的数量
```
```
   全局运行队列
   ---------------
   G - G - G - G
   ---------------
           ↑ ↓
   P1 ---- P2 ---- P3 ---- P4   (本地运行队列)
   |       |       |       |
   M1 ---- M2 ---- M3 ---- M4   (系统线程)
   |       |       |       |
   CPU1    CPU2    CPU3    CPU4  (物理核心)
```
```
P的数量决定了并行度
M会动态创建和销毁
G通过P在M上调度
工作窃取确保负载均衡
```
    - 这么说是不是M和CPU数量保存一致时是最高效的利用了CPU
```
M的数量应该是动态的
不是M越多越好
M的数量取决于：
P的数量（GOMAXPROCS）
系统调用阻塞情况
任务类型（CPU密集/IO密集）
```
```
   func example() {
       // 场景1: CPU密集型计算
       for i := 0; i < 4; i++ {
           go cpuIntensiveTask()  // 只需要4个M就够了
       }
       
       // 场景2: IO密集型操作
       for i := 0; i < 100; i++ {
           go func() {
               // 读取文件
               ioutil.ReadFile("large_file.txt")  // 系统调用会阻塞M
               // 此时需要更多的M来处理其他G
           }()
       }
   }
```
    - 那M阻塞的时候不会用到CPU吗？是不是M阻塞的时候在用其他硬件：磁盘，网络？
```
M阻塞时确实不使用CPU
具体的IO操作由相应的硬件控制器处理
通过中断机制通知操作完成
M在此期间完全让出CPU资源
```
```系统调用阻塞的类型
   // 1. IO阻塞（磁盘操作）
   func ioBlock() {
       file, _ := os.Open("large_file.txt")
       data := make([]byte, 1024)
       file.Read(data)  // M阻塞，等待磁盘IO完成
       // 此时：
       // - CPU：不使用
       // - 磁盘控制器：正在工作
   }
   
   // 2. 网络阻塞
   func networkBlock() {
       resp, _ := http.Get("http://example.com")
       // M阻塞，等待网络响应
       // 此时：
       // - CPU：不使用
       // - 网卡：正在工作
   }
```
```硬件交互过程：
   系统调用过程：
   
   用户态 --> 内核态 --> 设备驱动 --> 硬件操作 --> 中断返回 --> 内核态 --> 用户态
   
   M的状态：
   运行 --> 阻塞(不占用CPU) --> 等待中断 --> 被唤醒 --> 继续运行
```
```
   func example() {
       // 1. 磁盘IO操作
       go func() {
           file, _ := os.Open("file.txt")
           data := make([]byte, 1024)
           
           // 这里发生的事情：
           file.Read(data)  // 1. M1进入阻塞
                           // 2. P解绑定M1
                           // 3. P寻找/创建新的M2继续执行其他G
                           // 4. M1等待磁盘控制器的中断
                           // 5. M1不占用CPU
       }()
       
       // 2. 网络操作
       go func() {
           conn, _ := net.Dial("tcp", "example.com:80")
           
           // 这里发生的事情：
           data := make([]byte, 1024)
           conn.Read(data)  // 1. M2进入阻塞
                           // 2. P解绑定M2
                           // 3. P寻找/创建新的M3继续执行其他G
                           // 4. M2等待网卡的中断
                           // 5. M2不占用CPU
       }()
   }
```
    - 可以理解为阻塞时M让出了CPU然后交给其他硬件处理？那在多个阻塞时这些硬件是如何处理的？按顺序处理？
```
不同的硬件控制器可以并行工作
每个控制器都有自己的请求队列
现代硬件支持队列级并行
DMA减少CPU参与
```
``` 实际并行处理示例
   func parallelIO() {
       // 1. 磁盘IO
       go func() {
           file.Read(data)  // 磁盘控制器处理
           // 磁盘控制器可能同时处理其他请求
       }()
       
       // 2. 网络IO
       go func() {
           conn.Read(netData)  // 网络控制器处理
           // 网络控制器可能同时处理其他请求
       }()
       
       // 3. 不同控制器之间完全并行
       // 磁盘IO和网络IO可以同时进行
   }
```
## Goroutine是如何调度的?
多级队列调度
工作窃取平衡负载
抢占式调度保证公平
本地队列优先，减少锁竞争
``` 调度时机
   // Goroutine在以下情况下会触发调度
   func schedulePoints() {
       // 1. 系统调用
       syscall.Read(...)
       
       // 2. 通道操作
       ch <- data  // 发送阻塞
       <-ch       // 接收阻塞
       
       // 3. 垃圾回收
       runtime.GC()
       
       // 4. Go语句
       go func() {}()
       
       // 5. 内存同步访问
       mutex.Lock()
       
       // 6. 主动让出
       runtime.Gosched()
   }
```
```调度器的工作流程
   // 简化的调度循环
   func schedule() {
       for {
           // 1. 获取可运行的G
           g := findRunnable()
           
           // 2. 执行顺序：
           // - 本地队列
           // - 全局队列
           // - 网络轮询器
           // - 窃取工作
           
           // 3. 执行G
           execute(g)
       }
   }
```

## 大规模Goroutine的调度会产生什么影响?
内存压力增加
调度开销上升
资源竞争加剧
GC负担加重
runtime.GOMAXPROCS的作用是什么?

## Goroutin如何优雅地处理各种阻塞场景：
- IO阻塞（文件操作）
- 网络阻塞（HTTP请求）
- 锁阻塞（互斥锁）
- 睡眠阻塞（time.Sleep

## 为什么要避免过多的Goroutine切换?
## 为什么要注意内存分配和垃圾回收?
## 如何更好地利用CPU缓存?






















