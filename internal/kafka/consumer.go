package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	cfg "github.com/lyb88999/PortScan/internal/config"
	"github.com/lyb88999/PortScan/internal/models"
	rds "github.com/lyb88999/PortScan/internal/redis"
	"github.com/lyb88999/PortScan/internal/scanner"
	"strconv"
)

type ConsumerGroup struct {
	consumerGroup  sarama.ConsumerGroup
	rawTopic       string
	processedTopic string
	cfg            *cfg.Config
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
	// 获取一个Redis实例
	rdb := rds.GetRedisClient(cg.cfg)
	var ms = scanner.NewMasscanScanner(rdb)
	p, err := NewSyncProducer([]string{cg.cfg.KafkaHost}, cg.processedTopic)
	if err != nil {
		return fmt.Errorf("failed to new sync producer: %s", err)
	}
	for msg := range claim.Messages() {
		fmt.Printf("[consumer] topic:%q partition:%d offset:%d\n", msg.Topic, msg.Partition, msg.Offset)
		// 标记消息已被消费 内部会更新 consumer offset
		session.MarkMessage(msg, "")
		var data models.Data
		err := json.Unmarshal(msg.Value, &data)
		if err != nil {
			return fmt.Errorf("failed to unmarshal the data: %s", err)
		}
		scanOptions := scanner.ScanOptions{
			IP:        data.IP,
			Port:      data.Port,
			BandWidth: strconv.Itoa(data.Bandwidth),
		}
		scanResult, err := ms.Scan(scanOptions)
		if err != nil {
			return fmt.Errorf("failed to use masscan to do port scan: %s", err)
		}
		fmt.Println("masscan task is done")
		// 生产者向processedTopic发送处理好的消息
		for _, result := range scanResult {
			err = p.Send(result)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cg *ConsumerGroup) Close() error {
	return cg.consumerGroup.Close()
}

func NewConsumerGroup(brokers []string, groupID string, rawTopic string, processedTopic string, c *cfg.Config) (*ConsumerGroup, error) {
	config := sarama.NewConfig()
	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}
	return &ConsumerGroup{
		consumerGroup:  consumerGroup,
		rawTopic:       rawTopic,
		processedTopic: processedTopic,
		cfg:            c,
	}, nil
}

func (cg *ConsumerGroup) Consume(ctx context.Context) error {
	for {
		if err := cg.consumerGroup.Consume(ctx, []string{cg.rawTopic}, cg); err != nil {
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
