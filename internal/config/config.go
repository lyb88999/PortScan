package config

import (
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	RedisAddr      string `mapstructure:"REDIS_ADDR"`
	RedisPassword  string `mapstructure:"REDIS_PASSWORD"`
	RedisDB        int    `mapstructure:"REDIS_DB"`
	KafkaHost      string `mapstructure:"KAFKA_HOST"`
	InTopic        string `mapstructure:"INTOPIC"`
	ResultTopic    string `mapstructure:"RESULTTOPIC"`
	ProcessedTopic string `mapstructure:"PROCESSEDTOPIC"`
}

func LoadConfig(path string) (config *Config, err error) {
	if os.Getenv("CONFIG_PATH") != "" {
		path = os.Getenv("CONFIG_PATH")
	}
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
