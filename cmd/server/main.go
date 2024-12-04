package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lyb88999/PortScan/internal/config"
	"github.com/lyb88999/PortScan/internal/kafka"
	"github.com/lyb88999/PortScan/internal/redis"
	"github.com/lyb88999/PortScan/internal/scanner"
	"github.com/lyb88999/PortScan/internal/worker"
)

var cfg *config.Config

func init() {
	var err error
	// cfg, err = config.LoadConfig(".")
	cfg, err = config.LoadConfigFromExecutable()
	if err != nil {
		fmt.Println("failed to load config: ", err)
		os.Exit(-1)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("Received shutdown signal")
		cancel()
	}()

	// 传入Redis
	redisCli := redis.GetRedisClient(cfg)

	// 传入Kafka（生产者、消费者）
	producer, err := kafka.NewSyncProducer([]string{cfg.KafkaHost}, cfg.ProcessedTopic)
	if err != nil {
		fmt.Println("failed to new sync producer: ", err)
		return
	}

	groupID := "scanner_group"
	cg, err := kafka.NewConsumerGroup([]string{cfg.KafkaHost}, groupID, cfg.InTopic, cfg)
	if err != nil {
		fmt.Println("failed to new consumerGroup: ", err)
		return
	}
	// 传入PortScanner
	ps := scanner.NewMasscanScanner()

	// 传入worker
	w := worker.NewWorker(ctx, redisCli, producer, cg, ps)

	// 延迟调用worker的Stop方法
	defer func() {
		if err := w.Stop(); err != nil {
			fmt.Println("failed to stop worker: ", err)
		}
	}()

	// 调用worker的Run方法（从Kafka中读取出任务，扫描然后将进度写入Redis，将扫描结果发到Kafka）
	w.Run()
}
