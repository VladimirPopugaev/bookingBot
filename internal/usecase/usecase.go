package usecase

import (
	"booking_bot/internal/domain"
	"context"

	"github.com/rs/zerolog"
)

type Usecase struct {
	telegramRepo   domain.TelegramRepository
	siteWorkerRepo domain.SiteWorkerRepository
	siteParserRepo domain.SiteParserRepository
	databaseRepo   domain.DatabaseRepository
	log            zerolog.Logger

	cancelFunc context.CancelFunc
}

func New(
	telegramRepo domain.TelegramRepository,
	siteWorkerRepo domain.SiteWorkerRepository,
	siteParserRepo domain.SiteParserRepository,
	databaseRepo domain.DatabaseRepository,
	log zerolog.Logger,
) (domain.Usecase, error) {
	_, cancel := context.WithCancel(context.Background())

	uc := &Usecase{
		telegramRepo:   telegramRepo,
		siteWorkerRepo: siteWorkerRepo,
		siteParserRepo: siteParserRepo,
		databaseRepo:   databaseRepo,
		log:            log,
		cancelFunc:     cancel,
	}

	return uc, nil
}

func (uc *Usecase) Close() {
	uc.log.Trace().Msg("Usecase closing...")
	uc.cancelFunc()

	uc.telegramRepo.Close()
	uc.siteWorkerRepo.Close()
	uc.siteParserRepo.Close()
	uc.databaseRepo.Close()

	uc.log.Trace().Msg("Usecase closed successfully")
}
