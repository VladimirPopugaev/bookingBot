package tools

import (
	"errors"
	"fmt"
	"os"

	"booking_bot/internal/domain"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

func ParseConfig(configPath string, log zerolog.Logger) (*domain.Config, error) {
	if configPath == "" {
		return nil, errors.New("config path is empty")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file %q: %w", configPath, err)
	}

	var cfg domain.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse YAML from %q: %w", configPath, err)
	}

	return &cfg, nil
}
