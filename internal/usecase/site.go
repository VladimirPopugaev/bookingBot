package usecase

import (
	"booking_bot/internal/domain"
	"context"
	"net/url"
	"strings"
)

func (uc *Usecase) AnalyzeSite(ctx context.Context, rawURL string) (*domain.SiteInfo, error) {
	log := uc.log.With().Str("method", "AnalyzeSite").Str("url", rawURL).Logger()

	trimmedURL := strings.TrimSpace(rawURL)
	if trimmedURL == "" {
		log.Error().Msg("Analyze site url is empty")
		return nil, domain.ErrEmptyParameter
	}

	parsedURL, err := url.ParseRequestURI(trimmedURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		log.Error().Err(err).Str("url", trimmedURL).Msg("Analyze site url is invalid")
		return nil, domain.ErrURLParse
	}

	htmlStruct, err := uc.siteWorkerRepo.FetchSiteStruct(ctx, parsedURL.String())
	if err != nil {
		log.Error().Err(err).Str("url", trimmedURL).Msg("Failed to fetch site structure")
		return nil, domain.ErrCollectStruct
	}

	siteInfo, err := uc.siteWorkerRepo.ParseSiteStruct(ctx, strings.NewReader(htmlStruct))
	if err != nil {
		log.Error().Err(err).Str("url", trimmedURL).Msg("Failed to parse site structure")
		return nil, domain.ErrParseStruct
	}

	return siteInfo, nil
}
