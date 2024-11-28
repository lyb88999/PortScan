package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	cfg "github.com/lyb88999/PortScan/internal/config"
	"github.com/lyb88999/PortScan/internal/models"
	"github.com/lyb88999/PortScan/internal/scanner"
	"github.com/redis/go-redis/v9"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type ConsumerGroup struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	cfg           *cfg.Config
}

func (cg *ConsumerGroup) Setup(session sarama.ConsumerGroupSession) error {
	fmt.Println("Port scan service is started")
	return nil
}

func (cg *ConsumerGroup) Cleanup(session sarama.ConsumerGroupSession) error {
	fmt.Println("Port scan service is down")
	return nil
}

func (cg *ConsumerGroup) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("[consumer] topic:%q partition:%d offset:%d\n", msg.Topic, msg.Partition, msg.Offset)
		// 标记消息已被消费 内部会更新 consumer offset
		session.MarkMessage(msg, "")
		var data models.Data
		err := json.Unmarshal(msg.Value, &data)
		if err != nil {
			return fmt.Errorf("failed to unmarshal the data: %s", err)
		}
		rdb := redis.NewClient(&redis.Options{
			Addr:     cg.cfg.RedisAddr,
			Password: cg.cfg.RedisPassword,
			DB:       cg.cfg.RedisDB,
		})
		var ms = scanner.NewMasscanScanner(rdb)
		scanOptions := scanner.ScanOptions{
			IP:        data.IP,
			Port:      data.Port,
			BandWidth: strconv.Itoa(data.Bandwidth),
		}
		scanResult, err := ms.Scan(scanOptions)
		if err != nil {
			return fmt.Errorf("failed to use masscan to do port scan: %s", err)
		}
		fmt.Println(scanResult)
	}
	return nil
}

func NewConsumerGroup(brokers []string, groupID string, topic string) (*ConsumerGroup, error) {
	// 加载配置文件
	c, err := cfg.LoadConfig("../..")
	if err != nil {
		return nil, err
	}
	config := sarama.NewConfig()
	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}
	return &ConsumerGroup{
		consumerGroup: consumerGroup,
		topic:         topic,
		cfg:           c,
	}, nil
}

func (cg *ConsumerGroup) Consume(ctx context.Context) error {
	// 处理中断信号
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("context canceled, stopping consumption")
				return
			case <-signals:
				fmt.Println("signal received, stopping consumption")
				return
			}
		}
	}()

	// 启动消费者组
	for {
		if err := cg.consumerGroup.Consume(ctx, []string{cg.topic}, cg); err != nil {
			return fmt.Errorf("error from consumer: %v", err)
		}
		// 检查上下文是否已取消
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// 重新启动消费者组
		err := cg.consumerGroup.Close()
		if err != nil {
			return err
		}
	}
}
