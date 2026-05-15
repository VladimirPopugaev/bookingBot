package postgres

import (
	"booking_bot/internal/domain"
	"context"
	"errors"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

func TestBuildPoolConfig(t *testing.T) {
	t.Parallel()

	validCfg := domain.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Name:     "bookingbot",
		User:     "bookingbot",
		Password: "bookingbot",
		SSLMode:  "disable",
	}

	tests := []struct {
		name    string
		cfg     domain.DatabaseConfig
		wantErr error
	}{
		{
			name:    "empty host",
			cfg:     domain.DatabaseConfig{},
			wantErr: domain.ErrEmptyParameter,
		},
		{
			name: "invalid port",
			cfg: domain.DatabaseConfig{
				Host:     validCfg.Host,
				Port:     0,
				Name:     validCfg.Name,
				User:     validCfg.User,
				Password: validCfg.Password,
				SSLMode:  validCfg.SSLMode,
			},
			wantErr: domain.ErrInvalidParameter,
		},
		{
			name: "empty database name",
			cfg: domain.DatabaseConfig{
				Host:     validCfg.Host,
				Port:     validCfg.Port,
				Name:     "",
				User:     validCfg.User,
				Password: validCfg.Password,
				SSLMode:  validCfg.SSLMode,
			},
			wantErr: domain.ErrEmptyParameter,
		},
		{
			name: "empty ssl mode uses default",
			cfg: domain.DatabaseConfig{
				Host:     validCfg.Host,
				Port:     validCfg.Port,
				Name:     validCfg.Name,
				User:     validCfg.User,
				Password: validCfg.Password,
				SSLMode:  "",
			},
		},
		{
			name: "valid config",
			cfg:  validCfg,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			poolConfig, err := buildPoolConfig(tt.cfg, zerolog.Nop())
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if tt.wantErr != nil {
				if poolConfig != nil {
					t.Fatal("expected nil pool config on error")
				}

				return
			}

			if poolConfig == nil {
				t.Fatal("expected non-nil pool config")
			}

			if poolConfig.ConnConfig.Host != tt.cfg.Host {
				t.Fatalf("expected host %q, got %q", tt.cfg.Host, poolConfig.ConnConfig.Host)
			}

			if poolConfig.ConnConfig.TLSConfig != nil {
				t.Fatal("expected sslmode=disable to produce nil TLS config")
			}
		})
	}
}

func TestManagerBuilderUsesDollarPlaceholders(t *testing.T) {
	t.Parallel()

	manager := &repository{
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}

	query, args, err := manager.Builder().
		Select("id").
		From("site_analyses").
		Where(sq.Eq{"url": "https://example.com'}; DROP TABLE users; --"}).
		ToSql()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if query != "SELECT id FROM site_analyses WHERE url = $1" {
		t.Fatalf("expected dollar placeholders, got %q", query)
	}

	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
}

func TestManagerBeginTxNilPool(t *testing.T) {
	t.Parallel()

	manager := &repository{}

	tx, err := manager.BeginTx(t.Context(), pgx.TxOptions{})
	if err == nil {
		t.Fatal("expected error for nil pool")
	}

	if tx != nil {
		t.Fatal("expected nil tx when pool is nil")
	}
}

func TestRepositoryWithTxNilPool(t *testing.T) {
	t.Parallel()

	repo := &repository{}

	err := repo.WithTx(t.Context(), func(ctx context.Context, tx domain.Tx) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for nil pool")
	}
}

func TestRepositoryImplementsDatabaseRepository(t *testing.T) {
	t.Parallel()

	var _ domain.DatabaseRepository = (*repository)(nil)
}
