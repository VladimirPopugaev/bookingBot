package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	httpdelivery "booking_bot/internal/delivery/http"
	"booking_bot/internal/domain"
	"booking_bot/internal/repository/postgres"
	siteWorkerRepository "booking_bot/internal/repository/site"
	telegramRepository "booking_bot/internal/repository/telegram"
	"booking_bot/internal/usecase"
)

type App struct {
	log        zerolog.Logger
	uc         domain.Usecase
	httpServer *http.Server
}

func New(cfg *domain.Config, port int, log zerolog.Logger) (*App, error) {
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
		Timeout:            time.Duration(cfg.SiteConfig.RequestTimeout) * time.Second,
		MonitoringInterval: time.Duration(cfg.SiteConfig.MonitoringInterval) * time.Second,
	}
	siteWorkerRepo, err := siteWorkerRepository.New(siteWorkerCfg, log)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create site worker repository")
		return nil, fmt.Errorf("create site worker repository: %w", err)
	}
	log.Trace().Msg("Site worker repository created successfully")

	// TODO: добавлять его в sql-репозитории как QueryExecutor
	databaseRepo, err := postgres.New(cfg.Database, log)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create postgres repository")
		return nil, fmt.Errorf("create postgres repository: %w", err)
	}
	
	log.Trace().Msg("Postgres repository created successfully")

	uc, err := usecase.New(telegramRepo, siteWorkerRepo, databaseRepo, log)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create usecase")
		return nil, fmt.Errorf("create usecase: %w", err)
	}

	router := httpdelivery.NewRouter(uc, &log)
	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, port)
	server := httpdelivery.NewServer(addr, router)

	return &App{
		log:        log,
		uc:         uc,
		httpServer: server,
	}, nil
}

func (a *App) Start(cfg domain.HTTPConfig) {
	a.log.Trace().Msg("Starting app...")

	go func() {
		var err error
		if cfg.TLSEnabled && cfg.CertFile != "" && cfg.KeyFile != "" {
			a.log.Info().Msg("TLS is enabled, starting HTTPS server")
			a.log.Info().Str("addr", a.httpServer.Addr).Msg("Starting HTTPs server")
			err = a.httpServer.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
		} else {
			a.log.Info().Msg("TLS is disabled, starting HTTP server")
			a.log.Info().Str("addr", a.httpServer.Addr).Msg("Starting HTTP server")
			err = a.httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			a.log.Error().Err(err).Str("addr", a.httpServer.Addr).Msg("HTTP server stopped with error")
		}
	}()
}

func (a *App) Close() {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			a.log.Error().Err(err).Msg("Failed to shutdown HTTP server")
		}
	}

	a.uc.Close()
}
