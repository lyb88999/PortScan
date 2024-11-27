package kafka

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewSyncProducer(t *testing.T) {
	// 初始化生产者
	topic := "test"
	p, err := NewSyncProducer([]string{"47.93.190.244:9092"}, topic)
	require.NoError(t, err)
	require.Equal(t, p.topic, topic)
	require.NotEmpty(t, p.producer)
}

func TestProducerSend(t *testing.T) {
	// 初始化生产者
	topic := "test"
	p, err := NewSyncProducer([]string{"47.93.190.244:9092"}, topic)
	require.NoError(t, err)
	// 生产者发送消息
	err = p.Send("Hello Kafka")
	require.NoError(t, err)
}
