package kafka

import (
	"context"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
	cfg "github.com/lyb88999/PortScan/internal/config"
)

type ConsumerGroup struct {
	consumerGroup sarama.ConsumerGroup
	rawTopic      string
	cfg           *cfg.Config
	OutputChan    chan []byte
	ErrorChan     chan error
	done          chan struct{}
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
		select {
		case cg.OutputChan <- msg.Value:
			fmt.Printf("[consumer] topic:%q partition:%d offset:%d\n", msg.Topic, msg.Partition, msg.Offset)
			session.MarkMessage(msg, "")
		case <-session.Context().Done():
			return nil
		}
	}
	return nil
}

func (cg *ConsumerGroup) Close() error {
	close(cg.done) // 触发关闭信号
	if err := cg.consumerGroup.Close(); err != nil {
		return fmt.Errorf("failed to close consumer group: %w", err)
	}
	return nil
}

func NewConsumerGroup(brokers []string, groupID string, rawTopic string, c *cfg.Config) (*ConsumerGroup, error) {
	config := sarama.NewConfig()
	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}
	return &ConsumerGroup{
		consumerGroup: consumerGroup,
		rawTopic:      rawTopic,
		cfg:           c,
		OutputChan:    make(chan []byte, 1000),
		ErrorChan:     make(chan error, 1000),
		done:          make(chan struct{}),
	}, nil
}

func (cg *ConsumerGroup) Consume(ctx context.Context) error {
	defer close(cg.OutputChan)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-cg.done:
			return nil
		default:
			if err := cg.consumerGroup.Consume(ctx, []string{cg.rawTopic}, cg); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return nil
				}
				fmt.Printf("Error from consumer: %v, will retry...", err)
				continue
			}
		}
	}
}
