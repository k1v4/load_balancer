package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

// Config представляет основную структуру конфигурации
type Config struct {
	Port     string   `yaml:"port"`
	Backends []Client `yaml:"backends"`
}

// Client представляет параметры клиента
type Client struct {
	URL        string `yaml:"url"`
	BucketSize int    `yaml:"bucket_size"`
	RefillRate int    `yaml:"refill_rate"`
}

// GetConfig загружает конфигурацию из файла
func GetConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("config.yaml", cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return cfg, nil
}
