package kafka

import (
	"github.com/lyb88999/PortScan/internal/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewConsumerGroup(t *testing.T) {
	rawTopic := "test"
	processedTopic := "test2"
	groupID := "1"
	cfg, err := config.LoadConfig("../..")
	require.NoError(t, err)
	cg, err := NewConsumerGroup([]string{"47.93.190.244:9092"}, groupID, rawTopic, processedTopic, cfg)
	require.NoError(t, err)
	require.Equal(t, cg.rawTopic, rawTopic)
	require.Equal(t, cg.processedTopic, processedTopic)
	require.NotEmpty(t, cg.consumerGroup)
}
