package redis

import (
	"github.com/lyb88999/PortScan/internal/config"
	"github.com/redis/go-redis/v9"
	"sync"
)

var (
	redisClient *redis.Client
	once        sync.Once
)

func GetRedisClient(config *config.Config) *redis.Client {
	once.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     config.RedisAddr,
			Password: config.RedisPassword,
			DB:       config.RedisDB,
		})
	})
	return redisClient
}
