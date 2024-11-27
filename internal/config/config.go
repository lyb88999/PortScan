package config

import "github.com/spf13/viper"

type Config struct {
	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`
	KafkaHost     string `mapstructure:"KAFKA_HOST"`
	InTopic       string `mapstructure:"INTOPIC"`
	ResultTopic   string `mapstructure:"RESULTTOPIC"`
}

func LoadConfig(path string) (config *Config, err error) {
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
