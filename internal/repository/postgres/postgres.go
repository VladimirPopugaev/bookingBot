package postgres

import (
	"booking_bot/internal/domain"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

const (
	defaultConnectTimeout = 5 * time.Second
)

type repository struct {
	pool    *pgxpool.Pool
	builder sq.StatementBuilderType
	log     zerolog.Logger
}

func New(cfg domain.DatabaseConfig, logger zerolog.Logger) (domain.DatabaseRepository, error) {
	log := logger.With().Str("repository", "postgres").Logger()

	poolConfig, err := buildPoolConfig(cfg, log)
	if err != nil {
		log.Error().Err(err).Msg("Build postgres pool config failed")
		return nil, fmt.Errorf("build postgres pool config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Error().Err(err).Msg("Create postgres pool failed")
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Error().Err(err).Msg("Postgres ping failed")
		return nil, err
	}

	log.Trace().Msg("Postgres repository initialized successfully")

	return &repository{
		pool:    pool,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		log:     log,
	}, nil
}

func buildPoolConfig(cfg domain.DatabaseConfig, log zerolog.Logger) (*pgxpool.Config, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		log.Error().Msg("Postgres host is empty")
		return nil, domain.ErrEmptyParameter
	}

	if cfg.Port <= 0 {
		log.Error().Int("port", cfg.Port).Msg("Postgres port is invalid")
		return nil, domain.ErrInvalidParameter
	}

	if strings.TrimSpace(cfg.Name) == "" {
		log.Error().Msg("Postgres database name is empty")
		return nil, domain.ErrEmptyParameter
	}

	if strings.TrimSpace(cfg.User) == "" {
		log.Error().Msg("Postgres user is empty")
		return nil, domain.ErrEmptyParameter
	}

	if strings.TrimSpace(cfg.Password) == "" {
		log.Error().Msg("Postgres password is empty")
		return nil, domain.ErrEmptyParameter
	}

	if strings.TrimSpace(cfg.SSLMode) == "" {
		log.Warn().Msg("Postgres ssl mode is empty. Using disable")
		cfg.SSLMode = "disable"
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Error().Err(err).Msg("Parse postgres pool config failed")
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	poolConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout

	return poolConfig, nil
}

func (r *repository) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	if r == nil || r.pool == nil {
		return pgconn.CommandTag{}, errors.New("postgres pool is nil")
	}

	return r.pool.Exec(ctx, sql, arguments...)
}

func (r *repository) Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error) {
	if r == nil || r.pool == nil {
		return nil, errors.New("postgres pool is nil")
	}

	return r.pool.Query(ctx, sql, arguments...)
}

func (r *repository) QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row {
	if r == nil || r.pool == nil {
		r.log.Error().Str("query", sql).Msg("Postgres query row failed: pool is nil")
		return nil
	}

	return r.pool.QueryRow(ctx, sql, arguments...)
}

func (r *repository) Pool() *pgxpool.Pool {
	return r.pool
}

func (r *repository) Builder() sq.StatementBuilderType {
	return r.builder
}

func (r *repository) BeginTx(ctx context.Context, opts pgx.TxOptions) (domain.Tx, error) {
	if r == nil || r.pool == nil {
		return nil, errors.New("postgres pool is nil")
	}

	tx, err := r.pool.BeginTx(ctx, opts)
	if err != nil {
		r.log.Error().Err(err).Msg("Begin postgres transaction failed")
		return nil, err
	}

	return tx, nil
}

func (r *repository) WithTx(ctx context.Context, fn func(context.Context, domain.Tx) error) error {
	if r == nil || r.pool == nil {
		return errors.New("postgres pool is nil")
	}

	if fn == nil {
		return errors.New("transaction callback is nil")
	}

	tx, err := r.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		r.log.Error().Err(err).Msg("Begin postgres transaction for callback failed")
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err = fn(ctx, tx); err != nil {
		r.log.Error().Err(err).Msg("Postgres transaction callback failed")
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		r.log.Error().Err(err).Msg("Commit postgres transaction failed")
		return err
	}

	return nil
}

func (r *repository) Ping(ctx context.Context) error {
	if err := r.pool.Ping(ctx); err != nil {
		r.log.Error().Err(err).Msg("Postgres ping failed")
		return err
	}

	return nil
}

func (r *repository) Close() {
	r.pool.Close()
	r.log.Trace().Msg("Postgres repository closed")
}
