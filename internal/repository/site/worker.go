package site

import (
	"booking_bot/internal/domain"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
)

const (
	defaultTimeout            = 5 * time.Second
	defaultMonitoringInterval = 60 * time.Second
	textPreviewLimit          = 200
)

type worker struct {
	cfg       *Config
	collector *colly.Collector
	log       zerolog.Logger

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type Config struct {
	Timeout            time.Duration
	MonitoringInterval time.Duration
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

	// setup collector
	collector := colly.NewCollector()
	collector.AllowURLRevisit = true
	collector.SetRequestTimeout(cfg.Timeout)

	ctx, cancel := context.WithCancel(context.Background())

	log.Trace().Msg("Site worker repository initialized successfully")

	repo := &worker{
		cfg:       cfg,
		collector: collector,
		log:       log,
		ctx:       ctx,
		cancel:    cancel,
	}

	return repo, nil
}

func (c *Config) validate(log zerolog.Logger) error {
	if c.Timeout <= 0 {
		log.Warn().Dur("timeout", c.Timeout).Msg("Site worker timeout is negative. Using default timeout (5 seconds)")
		c.Timeout = defaultTimeout
	}

	if c.MonitoringInterval <= 0 {
		log.Warn().Dur("monitoring_interval", c.MonitoringInterval).Msg("Site worker monitoring interval is negative. Using default interval (60 seconds)")
		c.MonitoringInterval = defaultMonitoringInterval
	}

	return nil
}

func (w *worker) FetchSiteStruct(ctx context.Context, fetchUrl string) (string, error) {
	log := w.log.With().Str("method", "FetchSiteStruct").Str("target_url", fetchUrl).Logger()
	collector := w.collector.Clone()

	var html string
	var visitErr error

	collector.OnResponse(func(r *colly.Response) {
		html = string(r.Body)
		log.Info().
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

	if err := collector.Visit(fetchUrl); err != nil {
		log.Error().Err(err).Msg("Visit site html failed")
		return "", domain.ErrCollectStruct
	}

	if visitErr != nil {
		log.Error().Err(visitErr).Msg("Fetch site html finished with error")
		return "", domain.ErrCollectStruct
	}

	return html, nil
}

func (w *worker) backgroundMonitoring() {
	defer w.wg.Done()

	log := w.log.With().
		Str("method", "backgroundMonitoring").
		Dur("monitoring_interval", w.cfg.MonitoringInterval).
		Logger()

	log.Trace().Msg("background monitoring started")

	ticker := time.NewTicker(w.cfg.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			log.Trace().Msg("background monitoring stopped")
			return
		case <-ticker.C:
			_ = w.monitoringCheck(log)
		}
	}
}

func (w *worker) monitoringCheck(log zerolog.Logger) error {
	ctx, cancel := context.WithTimeout(w.ctx, w.cfg.Timeout)
	defer cancel()

	// TODO: delete monitoring and add monitorring to supervisor
	html, err := w.FetchSiteStruct(ctx, "")
	if err != nil {
		log.Error().Err(err).Msg("Site monitoring fetch site struct failed")
		return err
	}

	log.Info().
		Int("content_length", len(html)).
		Msg("Site monitoring iteration completed successfully")

	return nil
}

func (w *worker) Close() error {
	w.cancel()
	w.wg.Wait()
	w.log.Trace().Msg("Site worker repository closed")
	return nil
}
