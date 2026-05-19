package domain

import (
	"context"
	"io"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type Usecase interface {
	AnalyzeSite(ctx context.Context, rawURL string) (*SiteInfo, error)
	CheckSiteAvailability(ctx context.Context, rawURL string) (bool, error)
	Close()
}

type TelegramRepository interface {
	HealthCheck(ctx context.Context) error
	Close() error
}

type SiteWorkerRepository interface {
	// FetchSiteStruct will fetch the site structure and return it as a string (for now, it can be HTML or JSON)
	FetchSiteStruct(ctx context.Context, fetchURL string) (string, error)
	Close() error
}

type SiteParserRepository interface {
	ParseSiteStruct(ctx context.Context, htmlReader io.Reader) (*SiteInfo, error)
	IsAvailableToRegister(ctx context.Context, text string) (bool, error)
	Close() error
}

type DatabaseRepository interface {
	QueryExecutor

	BeginTx(ctx context.Context, opts pgx.TxOptions) (Tx, error)
	WithTx(ctx context.Context, fn func(context.Context, Tx) error) error
	Builder() squirrel.StatementBuilderType
	Ping(ctx context.Context) error
	Close()
}
