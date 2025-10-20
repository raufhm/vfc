package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig
	Worker WorkerConfig
	Queue  QueueConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

type WorkerConfig struct {
	Count int
}

type QueueConfig struct {
	BufferSize int
}

func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{
		Server: ServerConfig{
			Port:         viper.GetString("SERVER_PORT"),
			ReadTimeout:  viper.GetInt("SERVER_READ_TIMEOUT"),
			WriteTimeout: viper.GetInt("SERVER_WRITE_TIMEOUT"),
			IdleTimeout:  viper.GetInt("SERVER_IDLE_TIMEOUT"),
		},
		Worker: WorkerConfig{
			Count: viper.GetInt("WORKER_COUNT"),
		},
		Queue: QueueConfig{
			BufferSize: viper.GetInt("QUEUE_BUFFER_SIZE"),
		},
	}

	return config, nil
}
