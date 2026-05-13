package telegram

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"booking_bot/internal/domain"
)

const (
	defaultTimeout = 5 * time.Second
)

type repo struct {
	cfg    *Config
	client *http.Client
	log    zerolog.Logger
}

type Config struct {
	BotToken string
	BaseURL  string
	Timeout  time.Duration
}

func New(cfg *Config, log zerolog.Logger) (domain.TelegramRepository, error) {
	logger := log.With().Str("repository", "telegram").Logger()

	if cfg == nil {
		logger.Error().Msg("Telegram repository config is nil")
		return nil, fmt.Errorf("config is nil")
	}

	if err := cfg.validate(logger); err != nil {
		logger.Error().Err(err).Msg("validate config failed")
		return nil, fmt.Errorf("validate config: %w", err)
	}

	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	logger.Trace().Msg("Telegram repository initialized successfully")

	return &repo{
		cfg: &Config{
			BotToken: cfg.BotToken,
			BaseURL:  cfg.BaseURL,
			Timeout:  cfg.Timeout,
		},
		client: client,
		log:    logger,
	}, nil
}

func (c *Config) validate(log zerolog.Logger) error {
	if strings.TrimSpace(c.BotToken) == "" {
		log.Error().Msg("Telegram bot token is empty")
		return domain.ErrEmptyParameter
	}

	if strings.TrimSpace(c.BaseURL) == "" {
		log.Error().Msg("Telegram base url is empty")
		return domain.ErrEmptyParameter
	}

	parsedURL, err := url.ParseRequestURI(c.BaseURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		log.Error().Err(err).Str("base_url", c.BaseURL).Msg("Telegram base url is invalid")
		return domain.ErrUrlParse
	}

	if c.Timeout <= 0 {
		log.Warn().Int("timeout", int(c.Timeout.Seconds())).Msg("Telegram timeout not valid. Set to default value (5 seconds)")
		c.Timeout = defaultTimeout
	}

	return nil
}

func (r *repo) HealthCheck(ctx context.Context) error {
	return nil
}

func (r *repo) Close() error {
	// TODO: clear all resources if needed
	r.log.Trace().Msg("Telegram repository closed")
	return nil
}
