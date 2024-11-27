package kafka

import "github.com/IBM/sarama"

type ConsumerGroup struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
}

func NewConsumerGroup(brokers []string, groupID string, topic string) (*ConsumerGroup, error) {
	config := sarama.NewConfig()
	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}
	return &ConsumerGroup{
		consumerGroup: consumerGroup,
		topic:         topic,
	}, nil
}
