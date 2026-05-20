package usecase

import (
	"booking_bot/internal/domain"
	"context"
	"net/url"
	"strings"
)

func (uc *Usecase) CheckSiteForRegistration(ctx context.Context, siteURL string) (*domain.SiteInfo, error) {
	log := uc.log.With().Str("method", "CheckSiteForRegistration").Str("url", siteURL).Logger()

	trimmedURL := strings.TrimSpace(siteURL)
	if trimmedURL == "" {
		log.Error().Msg("Check site for registration url is empty")
		return nil, domain.ErrEmptyParameter
	}

	parsedURL, err := url.ParseRequestURI(trimmedURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		log.Error().Err(err).Str("url", trimmedURL).Msg("Check site for registration url is invalid")
		return nil, domain.ErrURLParse
	}

	htmlStruct, err := uc.siteWorkerRepo.FetchSiteStruct(ctx, parsedURL.String())
	if err != nil {
		log.Error().Err(err).Str("url", trimmedURL).Msg("Failed to fetch site structure")
		return nil, domain.ErrCollectStruct
	}

	siteInfo, err := uc.siteParserRepo.ParseSiteStruct(ctx, strings.NewReader(htmlStruct))
	if err != nil {
		log.Error().Err(err).Str("url", trimmedURL).Msg("Failed to parse site structure")
		return nil, domain.ErrParseStruct
	}

	isAvailable, err := uc.siteParserRepo.IsAvailableToRegister(ctx, htmlStruct)
	if err != nil {
		log.Error().Err(err).Str("url", trimmedURL).Msg("Failed to check site registration availability")
		return nil, domain.ErrParseStruct
	}

	siteInfo.URL = parsedURL.String()
	siteInfo.IsRegistrationAvailable = isAvailable

	return siteInfo, nil
}