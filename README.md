# PortScan
A PortScanner with masscan using Golang

需求:
1. 客户端程序：作为生产者，写一个命令行程序，往kafka里面推消息：ip、port、带宽 cmd/client/portscan

2. 服务端程序：作为消费者：从Kafka里面提取ip、port、带宽，调用masscan完成扫描，捕获标准输出的结果（传输层协议、IP、port），结果解析为json再推到kafka，以一条一条的数据往kafka里面推，masscan的标准错误stderr里面有扫描的进度，把扫描的进度写入到redis里面。需要支持两个独立的进程同时运行