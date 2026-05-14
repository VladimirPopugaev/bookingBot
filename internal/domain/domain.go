package domain

import (
	"context"
	"io"
)

type Usecase interface {
	Close()
}

type TelegramRepository interface {
	HealthCheck(ctx context.Context) error
	Close() error
}

type SiteWorkerRepository interface {
	// FetchSiteStruct will fetch the site structure and return it as a string (for now, it can be HTML or JSON)
	FetchSiteStruct(ctx context.Context) (string, error)
	ParseSiteStruct(ctx context.Context, htmlReader io.Reader) (*SiteInfo, error)
	Close() error
}
