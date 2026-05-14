package app

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"booking_bot/internal/domain"
	siteWorkerRepository "booking_bot/internal/repository/site"
	telegramRepository "booking_bot/internal/repository/telegram"
	"booking_bot/internal/usecase"
)

type App struct {
	logger zerolog.Logger
	uc     domain.Usecase
}

func New(cfg *domain.Config, log zerolog.Logger) (*App, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	// create telegram repository
	telegramCfg := &telegramRepository.Config{
		BotToken: cfg.Telegram.BotToken,
		BaseURL:  cfg.Telegram.BaseURL,
		Timeout:  time.Duration(cfg.Telegram.Timeout) * time.Second,
	}
	telegramRepo, err := telegramRepository.New(telegramCfg, log)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create telegram repository")
		return nil, fmt.Errorf("create telegram repository: %w", err)
	}
	log.Trace().Msg("Telegram repository created successfully")

	// create site worker repository
	siteWorkerCfg := &siteWorkerRepository.Config{
		TargetURL:          cfg.SiteConfig.TargetURL,
		Timeout:            time.Duration(cfg.SiteConfig.RequestTimeout) * time.Second,
		MonitoringInterval: time.Duration(cfg.SiteConfig.MonitoringInterval) * time.Second,
	}
	siteWorkerRepo, err := siteWorkerRepository.New(siteWorkerCfg, log)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create site worker repository")
		return nil, fmt.Errorf("create site worker repository: %w", err)
	}
	log.Trace().Msg("Site worker repository created successfully")

	uc, err := usecase.New(telegramRepo, siteWorkerRepo, log)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create usecase")
		return nil, fmt.Errorf("create usecase: %w", err)
	}

	return &App{
		logger: log,
		uc:     uc,
	}, nil
}

func (a *App) Close() {
	a.uc.Close()
}
