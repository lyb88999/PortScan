package scanner

import (
	"github.com/lyb88999/PortScan/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMasscanScanner(t *testing.T) {
	// 加载配置文件
	cfg, err := config.LoadConfig("../..")
	require.NoError(t, err)
	// 创建redis client
	// var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	var ms = NewMasscanScanner(rdb)
	scanOptions := ScanOptions{
		IP:        "47.93.190.244/20",
		Port:      8080,
		BandWidth: "10000",
	}
	scanResult, err := ms.Scan(scanOptions)
	require.NoError(t, err)
	t.Log(scanResult)
}
