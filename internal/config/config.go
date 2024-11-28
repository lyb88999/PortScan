package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
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

func GetExecutableDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %s", err)
	}
	return filepath.Dir(execPath), nil
}

func LoadConfigFromExecutable() (*Config, error) {
	execDir, err := GetExecutableDir()
	if err != nil {
		return nil, err
	}
	configPaths := []string{
		filepath.Join(execDir),
		filepath.Join(execDir, "configs"),
		"../..",
	}
	var lastErr error
	for _, path := range configPaths {
		config, err := LoadConfig(path)
		if err == nil {
			return config, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("failed to load config: %s", lastErr)
}
