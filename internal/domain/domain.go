package domain

import "context"

type Usecase interface {
	Close()
}

type TelegramRepository interface {
	HealthCheck(ctx context.Context) error
	Close() error
}
