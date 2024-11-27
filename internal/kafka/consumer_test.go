package kafka

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewConsumerGroup(t *testing.T) {
	topic := "test"
	groupID := "1"
	cg, err := NewConsumerGroup([]string{"47.93.190.244:9092"}, groupID, topic)
	require.NoError(t, err)
	require.Equal(t, cg.topic, topic)
	require.NotEmpty(t, cg.consumerGroup)
}
