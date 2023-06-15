# 基于goroutine的tcp压测工具

### 背景：没找到wrk tcp类型的压测工具，简单实现下

### 功能：并发压测并统计响应和请求速度

### 快速上手：

配置好压测目标

```
config.yaml
TargetIp: 127.0.0.1  目标ip
TargetPort: 6000 目标端口
CoroutineNum: 25000 协程数
Duration: 300 持续时间s
MaxTimeout: 200 单个请求超时ms
Cmd: "command=CmdCurlSo&url=http%3A%2F%2Fwww.baidu.com%2F&type=get&content=" 字符串
```

```
./wrk_tcp_tool 启动
```

返回结果

```
[wrk_tcp_tool]2023/06/13 19:17:48.857153 server.go:204: ----total req in 60 sencond 559144-----
[wrk_tcp_tool]2023/06/13 19:17:48.857179 server.go:205: QPms:9/ms  请求速率
[wrk_tcp_tool]2023/06/13 19:17:48.857188 server.go:206: QPS_out:9319/s 请求速率
[wrk_tcp_tool]2023/06/13 19:17:48.857196 server.go:207: TOTAL_REQ:559144 总请求数
[wrk_tcp_tool]2023/06/13 19:17:48.857204 server.go:208: TOTAL_AVG:710002776ms 响应总耗时
[wrk_tcp_tool]2023/06/13 19:17:48.857215 server.go:209: <100_RSP:17.34920%,num:97007  100ms内占比
[wrk_tcp_tool]2023/06/13 19:17:48.857223 server.go:210: <200_RSP:1.08863%,num:6087
[wrk_tcp_tool]2023/06/13 19:17:48.857233 server.go:211: <300_RSP:1.44614%,num:8086
[wrk_tcp_tool]2023/06/13 19:17:48.857241 server.go:212: AVG_RSP:1269ms 平均响应耗时
[wrk_tcp_tool]2023/06/13 19:17:48.857249 server.go:213: QPS_RSP:0/s
[wrk_tcp_tool]2023/06/13 19:17:48.857258 server.go:214: CoroutineNum:25000 协程数量
[wrk_tcp_tool]2023/06/13 19:17:48.857269 main.go:60: -----wrk_tcp_tool start run over:5074-----
```

QPS_out 可以反应对目标端口的施压流量，TOTAL_AVG 反应了面对施压流量服务器返回请求使用多少时间。
提高CoroutineNum不一定能够提高QPS_out ，但会使得TOTAL_AVG 时间进一步变长，以此来判断出不同时间分位下100ms,200ms,300ms内请求占比



### 基本原理：

利用go协程来实现大并发，启用目标个数协程，不断发送请求，达到压测目的

### 待优化:

1.单个请求超时控制：
服务响应超时后会一直阻塞单个协程，使得出口并发量不及预期。希望对单个请求设置超时，避免阻塞该协程。

2.响应处理：
目前对返回结果不敏感，不会记录异常数量

```
func (this *Server) RspHandler(Recv []byte)
该函数可以添加对rsp的处理来记录异常
```


