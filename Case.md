# 从0到生成上线

## 1.编写代码
> 以Golang为例

### 1-1.项目设计
- 模块分析

### 1-2.关系型数据库设计
表结构设计
- 考虑读写频率
- 主键设计

### 1-3.实现功能
语言特性：
- 并发：Goroutine

- 通信方式：Channel
CSP
- 锁：sync.Mutex

组件调用：
1.监听端口绑定链接
2.接收请求数据
3.处理数据
4.发起外包请求：Mysql，Redis
5.返回响应



### 1-4.引入组件
- 关系行数据库：Mysql，PSql
- 缓存类：Redis
- 文档类：Mongo
- 数据存储类：ES，Clickhouse
- 消息类：Kafka


## 2.上线


## 3.运维&监控


## 参考
https://blog.algomaster.io/p/30-system-design-concepts