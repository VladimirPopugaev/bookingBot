package storage

import (
	"booking_bot/internal/domain"

	"github.com/rs/zerolog"
)

type repository struct{}

func New(logger zerolog.Logger, postgresDB domain.QueryExecutor) (domain.SiteStorageRepository, error) {
	return nil, nil
}