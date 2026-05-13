package usecase

import (
	"booking_bot/internal/domain"
	"context"

	"github.com/rs/zerolog"
)

type Usecase struct {
	telegramRepo   domain.TelegramRepository
	siteWorkerRepo domain.SiteWorkerRepository
	log            zerolog.Logger

	cancelFunc context.CancelFunc
}

func New(
	telegramRepo domain.TelegramRepository,
	siteWorkerRepo domain.SiteWorkerRepository,
	log zerolog.Logger,
) (*Usecase, error) {
	_, cancel := context.WithCancel(context.Background())

	uc := &Usecase{
		telegramRepo:   telegramRepo,
		siteWorkerRepo: siteWorkerRepo,
		log:            log,
		cancelFunc:     cancel,
	}

	// TODO: удалить когда будет реализован полный функционал вызова метода
	htmlBody, err := uc.siteWorkerRepo.FetchSiteStruct(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Fetch site structure failed")
		return nil, domain.ErrCollectStruct
	}
	log.Trace().Str("site_struct", htmlBody).Msg("Site structure fetched successfully")

	return uc, nil
}

func (uc *Usecase) Close() {
	uc.log.Trace().Msg("Usecase closing...")
	uc.cancelFunc()

	uc.telegramRepo.Close()
	uc.siteWorkerRepo.Close()

	uc.log.Trace().Msg("Usecase closed successfully")
}
