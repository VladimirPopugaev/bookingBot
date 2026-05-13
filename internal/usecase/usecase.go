package usecase

import (
	"booking_bot/internal/domain"
	"context"

	"github.com/rs/zerolog"
)

type Usecase struct {
	telegramRepo domain.TelegramRepository
	log          zerolog.Logger

	cancelFunc context.CancelFunc
}

func New(telegramRepo domain.TelegramRepository, log zerolog.Logger) (*Usecase, error) {
	_, cancel := context.WithCancel(context.Background())

	return &Usecase{
		telegramRepo: telegramRepo,
		log:          log,
		cancelFunc:   cancel,
	}, nil
}

func (uc *Usecase) Close() {
	uc.log.Trace().Msg("Usecase closing...")
	uc.cancelFunc()
	uc.telegramRepo.Close()
}
