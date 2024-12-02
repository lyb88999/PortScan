## PortScan
一个基于Kafka的端口扫描系统
## 功能特性
* 基于masscan的高性能扫描
* 实时进度监控(Redis)
* JSON格式输出
## 系统架构
* 命令行客户端: 提交扫描任务
* Kafka: 消息队列，负责任务调度
* 扫描服务器: 消费Kafka消息，执行扫描任务，并将扫描结果推送到Kafka
* Redis: 存储扫描进度
## 前置要求
* Go
* Kafka
* Redis
* masscan
## 安装
```bash
# 克隆项目
git clone https://github.com/lyb88999/PortScan.git

# 安装依赖
go mod download

# 编译客户端
make client

# 编译服务端并运行
make server

# 使用Dockerfile构建镜像(macOS)
make docker
```

## 配置文件
创建`app.env`文件，补充以下内容
```
REDIS_ADDR="localhost:6379"
REDIS_PASSWORD=""
REDIS_DB=0
KAFKA_HOST="47.93.190.244:9092"
INTOPIC="portscan2"
PROCESSEDTOPIC="portscan_result"
```
## 使用方法
1. 启动扫描服务器:
```bash
make server
```
2. 提交扫描任务
```bash
./portscan masscan --ip "192.168.1.0/24" --port 80 --bandwidth 1000
```

参数说明
* `--ip`: 扫描的目标IP或IP范围
* `--port`: 扫描的目标端口
* `--bandwidth`: 扫描的带宽限制

## 代码结构
```
.
├── cmd/
│   ├── client/         # 客户端入口
│   ├── server/         # 服务端入口
│   └── masscan.go      # masscan 命令实现
├── internal/
│   ├── config/         # 配置管理
│   ├── kafka/          # Kafka 相关实现
│   ├── models/         # 数据模型
│   ├── redis/          # Redis 相关实现
│   └── scanner/        # 扫描器实现
└── config.yaml         # 配置文件
```

