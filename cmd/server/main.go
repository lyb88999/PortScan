package main

import (
	"context"
	"fmt"
	"github.com/lyb88999/PortScan/internal/config"
	"github.com/lyb88999/PortScan/internal/kafka"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var cfg *config.Config

func init() {
	var err error
	cfg, err = config.LoadConfig("../..")
	if err != nil {
		fmt.Println("failed to load config: ", err)
	}
}
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 捕获中断信号
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		fmt.Println("signal received, stopping consumption")
		cancel()
	}()

	// 创建消费者组
	cg, err := kafka.NewConsumerGroup([]string{cfg.KafkaHost}, "1", cfg.InTopic, cfg.ProcessedTopic, cfg)
	if err != nil {
		fmt.Println("failed to new consumerGroup: ", err)
		return
	}
	defer func(cg *kafka.ConsumerGroup) {
		err := cg.Close()
		if err != nil {
			fmt.Println("failed to close consumerGroup: ", err)
		}
	}(cg)

	// 开始消费
	err = cg.Consume(ctx)
	if err != nil {
		log.Fatalln("failed to consume: ", err)
	}
}
