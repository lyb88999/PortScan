package main

import (
	"context"
	"fmt"
	"github.com/lyb88999/PortScan/internal/config"
	"github.com/lyb88999/PortScan/internal/kafka"
	"log"
)

var cfg *config.Config

func init() {
	var err error
	cfg, err = config.LoadConfig(".")
	if err != nil {
		fmt.Println("failed to load config: ", err)
	}
}
func main() {
	cg, err := kafka.NewConsumerGroup([]string{cfg.KafkaHost}, "1", cfg.InTopic)
	if err != nil {
		fmt.Println("failed to new consumerGroup: ", err)
	}
	err = cg.Consume(context.Background())
	if err != nil {
		log.Fatalln("failed to consume: ", err)
	}
}
