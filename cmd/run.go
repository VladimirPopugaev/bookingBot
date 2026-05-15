package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"booking_bot/internal/app"
	"booking_bot/internal/domain"
	"booking_bot/tools"
)

var port int
var configPath string

var serveCmd = &cobra.Command{
	Use:   "run",
	Short: "Run booking bot service",
	Run: func(cmd *cobra.Command, args []string) {
		log := tools.NewLogger()

		cfg, err := tools.ParseConfig(configPath, log)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse config")
			os.Exit(1)
		}
		log.Trace().Msg("Config parsed successfully")

		level, err := zerolog.ParseLevel(cfg.LogLevel)
		if err != nil {
			log.Error().Err(err).Str("log_level", cfg.LogLevel).Msg("Invalid log level in config")
			os.Exit(1)
		}
		log = log.Level(level)
		log.Trace().Msg("Logger initialized")

		if err := ValidateConfig(cfg, port); err != nil {
			log.Error().Err(err).Int("port", port).Msg("Invalid runtime config")
			os.Exit(1)
		}

		application, err := app.New(cfg, port, log)
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize application")
			os.Exit(1)
		}

		application.Start(cfg.HTTP)
		log.Info().Msg("Success initialization")

		chExit := make(chan os.Signal, 1)
		defer close(chExit)

		signal.Notify(chExit, os.Interrupt, syscall.SIGTERM)

		sig := <-chExit
		log.Info().Str("signal", sig.String()).Msg("Shutdown signal received")

		application.Close()
		log.Info().Msg("Application stopped gracefully")
	},
}

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port for service startup")
	serveCmd.Flags().StringVarP(&configPath, "config", "c", "config.yml", "Path to YAML config file")

	rootCmd.AddCommand(serveCmd)
}

func ValidateConfig(cfg *domain.Config, port int) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	if port <= 0 {
		return fmt.Errorf("invalid port: must be greater than zero")
	}

	if strings.TrimSpace(cfg.HTTP.Host) == "" {
		cfg.HTTP.Host = "0.0.0.0"
	}

	if !cfg.HTTP.TLSEnabled {
		return nil
	}

	if strings.TrimSpace(cfg.HTTP.CertFile) == "" {
		return fmt.Errorf("http cert_file is required when tls_enabled=true")
	}

	if strings.TrimSpace(cfg.HTTP.KeyFile) == "" {
		return fmt.Errorf("http key_file is required when tls_enabled=true")
	}

	return nil
}
