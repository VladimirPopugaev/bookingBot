package siteworker

import (
	"booking_bot/internal/domain"
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
)

const (
	defaultTimeout = 5 * time.Second
)

// TODO: implement site worker that will check the site for availability
// TODO: implement site worker that will parse the site structure and parse info about it in SiteInfo struct

type worker struct {
	cfg       *Config
	collector *colly.Collector
	log       zerolog.Logger
}

type Config struct {
	TargetURL string
	Timeout   time.Duration
}

func New(cfg *Config, logger zerolog.Logger) (domain.SiteWorkerRepository, error) {
	log := logger.With().Str("repository", "site_worker").Logger()

	if cfg == nil {
		log.Error().Msg("Site worker config is nil")
		return nil, fmt.Errorf("config is nil")
	}

	if err := cfg.validate(log); err != nil {
		log.Error().Err(err).Msg("Validate site worker config failed")
		return nil, fmt.Errorf("validate config: %w", err)
	}

	collector := colly.NewCollector()
	collector.SetRequestTimeout(cfg.Timeout)

	log.Trace().Msg("Site worker repository initialized successfully")

	return &worker{
		cfg: &Config{
			TargetURL: cfg.TargetURL,
			Timeout:   cfg.Timeout,
		},
		collector: collector,
		log:       log,
	}, nil
}

func (c *Config) validate(log zerolog.Logger) error {
	if strings.TrimSpace(c.TargetURL) == "" {
		log.Error().Msg("Site worker target url is empty")
		return domain.ErrEmptyParameter
	}

	parsedURL, err := url.ParseRequestURI(c.TargetURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		log.Error().Err(err).Str("target_url", c.TargetURL).Msg("Site worker target url is invalid")
		return domain.ErrURLParse
	}

	if c.Timeout < 0 {
		log.Warn().Dur("timeout", c.Timeout).Msg("Site worker timeout is negative. Using default timeout (5 seconds)")
		c.Timeout = defaultTimeout
	}

	return nil
}

func (w *worker) FetchSiteStruct(ctx context.Context) (string, error) {
	log := w.log.With().Str("method", "FetchSiteStruct").Str("target_url", w.cfg.TargetURL).Logger()
	collector := w.collector.Clone()

	var html string
	var visitErr error

	collector.OnResponse(func(r *colly.Response) {
		html = string(r.Body)
		log.Info().
			Str("url", r.Request.URL.String()).
			Int("status_code", r.StatusCode).
			Int("content_length", len(r.Body)).
			Msg("Fetched site html successfully")
	})

	collector.OnError(func(r *colly.Response, collectErr error) {
		visitErr = collectErr
		logEvent := log.Error().Err(collectErr)
		if r != nil {
			logEvent = logEvent.Int("status_code", r.StatusCode)
		}

		logEvent.Msg("Fetch site html failed")
	})

	if err := collector.Visit(w.cfg.TargetURL); err != nil {
		log.Error().Err(err).Msg("Visit site html failed")
		return "", domain.ErrCollectStruct
	}

	if visitErr != nil {
		log.Error().Err(visitErr).Msg("Fetch site html finished with error")
		return "", domain.ErrCollectStruct
	}

	return html, nil
}

func (w *worker) ParseSiteStruct(ctx context.Context, siteStruct string) (*domain.SiteInfo, error) {
	// TODO: implement site structure parsing logic
	return nil, nil
}

func (w *worker) Close() error {
	w.log.Trace().Msg("Site worker repository closed")
	return nil
}
