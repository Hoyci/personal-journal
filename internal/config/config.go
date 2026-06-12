package config

import (
	"github.com/hoyci/personal-journal/internal/core"

	"github.com/spf13/viper"
)

type SourceConfig struct {
	Name     string
	URL      string
	Category string
	Priority core.Priority
}

type Config struct {
	Sources []SourceConfig
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
