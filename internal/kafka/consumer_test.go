package kafka

import (
	"testing"

	"github.com/lyb88999/PortScan/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNewConsumerGroup(t *testing.T) {
	rawTopic := "test"
	groupID := "1"
	cfg, err := config.LoadConfig("../..")
	require.NoError(t, err)
	cg, err := NewConsumerGroup([]string{"47.93.190.244:9092"}, groupID, rawTopic, cfg)
	require.NoError(t, err)
	require.Equal(t, cg.rawTopic, rawTopic)
	require.NotEmpty(t, cg.consumerGroup)
}
