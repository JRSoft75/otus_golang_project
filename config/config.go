package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoggerConfig представляет настройки логгера.
type LoggerConfig struct {
	Level string `yaml:"level"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// Config представляет основную структуру конфигурации сервиса.
type Config struct {
	Logger  LoggerConfig `yaml:"logger"`
	Server  ServerConfig `yaml:"server"`
	Storage struct {
		CacheSize            int    `yaml:"cacheSize"`
		CacheDir             string `yaml:"cacheDir"`
		DefaultImageQuality  int    `yaml:"defaultImageQuality"`
		MaxUploadedImageSize int    `yaml:"maxUploadedImageSize"` // in megabytes
		ReadTimeout          int    `yaml:"readTimeout"`
	} `yaml:"storage"`
}

func LoadConfig(filePath string) (*Config, error) {
	// Проверка существования файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", filePath)
	}

	// Чтение содержимого файла
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Парсинг YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
