package usecase

import (
	"booking_bot/internal/domain"
	"context"
	"strings"

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

	// TODO: remove
	htmlContent, err := siteWorkerRepo.FetchSiteStruct(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Fetch site struct failed")
		return nil, err
	}

	siteInfo, err := siteWorkerRepo.ParseSiteStruct(context.Background(), strings.NewReader(htmlContent))
	if err != nil {
		log.Error().Err(err).Msg("Parse site struct failed")
		return nil, err
	}
	log.Info().Interface("siteInfo", siteInfo).Msg("Site info parsed successfully")

	
	return uc, nil
}

func (uc *Usecase) Close() {
	uc.log.Trace().Msg("Usecase closing...")
	uc.cancelFunc()

	uc.telegramRepo.Close()
	uc.siteWorkerRepo.Close()

	uc.log.Trace().Msg("Usecase closed successfully")
}
