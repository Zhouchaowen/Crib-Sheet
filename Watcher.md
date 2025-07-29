## 链接详细查询
- 使用 netstat 命令
```
# 查看所有TCP连接
netstat -nat

# 查看更详细的连接信息，包括PID和程序名
netstat -natp

# 实时监控连接状态
netstat -nat | watch
```
- 使用 ss 命令
```
# 查看所有TCP连接
ss -t
watch -n 1 'ss -t'

# 查看详细连接信息
ss -tani

# 查看监听端口
ss -tlnp

# 查看与特定IP的所有连接
ss -tn dst 192.168.1.100

# 查看与特定IP的详细连接信息
ss -tni dst 192.168.1.100

# 查看来自特定IP的所有连接
ss -tn src 192.168.1.100

# 查看来自特定IP的详细连接信息
ss -tni src 192.168.1.100

# 查看特定目标端口的连接
ss -tn dst :80

# 查看特定源端口的连接
ss -tn src :80

# 查看特定端口（源或目标）的连接
ss -tn "sport = :80 or dport = :80"

# 查看特定IP和端口的连接
ss -tn dst 192.168.1.100:80

# 使用更复杂的过滤条件
ss -tn '( dst 192.168.1.100:80 or dst 192.168.1.100:443 )'

# 统计与特定IP的连接数
ss -tn dst 192.168.1.100 | wc -l

# 统计特定端口的连接数
ss -tn dst :80 | wc -l

# 常用选项说明：
-t：只显示TCP连接
-n：不解析服务名称
-i：显示详细的socket信息
-p：显示进程信息
-s：显示统计信息
```
## 内存占用详情
- 使用 ps 命令
```
# 按内存使用率排序
ps aux --sort=-%mem | head -n 11

# 格式化输出（更易读）
ps -eo pid,ppid,%mem,%cpu,cmd --sort=-%mem | head -n 11

# 显示完整命令
ps auxf --sort=-%mem | head -n 11

# 实时监控（每3秒更新一次）
watch -n 3 'ps aux --sort=-%mem | head -n 11'
```