package siteworker

import (
	"booking_bot/internal/domain"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cfg       *Config
		wantErr   error
		assertNew func(t *testing.T, repo any, cfg *Config)
	}{
		{
			name: "valid config",
			cfg: &Config{
				TargetURL: "https://example.com",
				Timeout:   5 * time.Second,
			},
			assertNew: func(t *testing.T, repo any, cfg *Config) {
				t.Helper()

				workerRepo, ok := repo.(*worker)
				if !ok {
					t.Fatalf("expected repository type *worker, got %T", repo)
				}

				if workerRepo.collector == nil {
					t.Fatal("expected collector to be initialized")
				}

				if workerRepo.cfg == cfg {
					t.Fatal("expected config to be copied, got the same pointer")
				}

				if workerRepo.cfg.TargetURL != cfg.TargetURL {
					t.Fatalf("expected target url %q, got %q", cfg.TargetURL, workerRepo.cfg.TargetURL)
				}

				if workerRepo.cfg.Timeout != cfg.Timeout {
					t.Fatalf("expected timeout %s, got %s", cfg.Timeout, workerRepo.cfg.Timeout)
				}
			},
		},
		{
			name: "empty target url",
			cfg: &Config{
				TargetURL: "",
				Timeout:   5 * time.Second,
			},
			wantErr: domain.ErrEmptyParameter,
		},
		{
			name: "blank target url",
			cfg: &Config{
				TargetURL: "   ",
				Timeout:   5 * time.Second,
			},
			wantErr: domain.ErrEmptyParameter,
		},
		{
			name: "invalid target url",
			cfg: &Config{
				TargetURL: "://broken-url",
				Timeout:   5 * time.Second,
			},
			wantErr: domain.ErrUrlParse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo, err := New(tt.cfg, zerolog.Nop())
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}

				if repo != nil {
					t.Fatalf("expected nil repository, got %T", repo)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if repo == nil {
				t.Fatal("expected repository to be initialized")
			}

			if tt.assertNew != nil {
				tt.assertNew(t, repo, tt.cfg)
			}
		})
	}
}

func TestWorker_FetchSiteStruct(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		const expectedHTML = "<html><body>ok</body></html>"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(expectedHTML))
		}))
		defer server.Close()

		repo, err := New(&Config{
			TargetURL: server.URL,
			Timeout:   5 * time.Second,
		}, zerolog.Nop())
		if err != nil {
			t.Fatalf("expected no error creating repository, got %v", err)
		}

		workerRepo, ok := repo.(*worker)
		if !ok {
			t.Fatalf("expected repository type *worker, got %T", repo)
		}

		html, err := workerRepo.FetchSiteStruct(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if html != expectedHTML {
			t.Fatalf("expected html %q, got %q", expectedHTML, html)
		}
	})

	t.Run("server unavailable", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		serverURL := server.URL
		server.Close()

		repo, err := New(&Config{
			TargetURL: serverURL,
			Timeout:   500 * time.Millisecond,
		}, zerolog.Nop())
		if err != nil {
			t.Fatalf("expected no error creating repository, got %v", err)
		}

		workerRepo, ok := repo.(*worker)
		if !ok {
			t.Fatalf("expected repository type *worker, got %T", repo)
		}

		html, err := workerRepo.FetchSiteStruct(context.Background())
		if !errors.Is(err, domain.ErrCollectStruct) {
			t.Fatalf("expected error %v, got %v", domain.ErrCollectStruct, err)
		}

		if html != "" {
			t.Fatalf("expected empty html, got %q", html)
		}
	})

	t.Run("server returns 500", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}))
		defer server.Close()

		repo, err := New(&Config{
			TargetURL: server.URL,
			Timeout:   5 * time.Second,
		}, zerolog.Nop())
		if err != nil {
			t.Fatalf("expected no error creating repository, got %v", err)
		}

		workerRepo, ok := repo.(*worker)
		if !ok {
			t.Fatalf("expected repository type *worker, got %T", repo)
		}

		html, err := workerRepo.FetchSiteStruct(context.Background())
		if !errors.Is(err, domain.ErrCollectStruct) {
			t.Fatalf("expected error %v, got %v", domain.ErrCollectStruct, err)
		}

		if html != "" {
			t.Fatalf("expected empty html, got %q", html)
		}
	})

	t.Run("request timeout exceeded", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<html><body>slow</body></html>"))
		}))
		defer server.Close()

		repo, err := New(&Config{
			TargetURL: server.URL,
			Timeout:   50 * time.Millisecond,
		}, zerolog.Nop())
		if err != nil {
			t.Fatalf("expected no error creating repository, got %v", err)
		}

		workerRepo, ok := repo.(*worker)
		if !ok {
			t.Fatalf("expected repository type *worker, got %T", repo)
		}

		html, err := workerRepo.FetchSiteStruct(context.Background())
		if !errors.Is(err, domain.ErrCollectStruct) {
			t.Fatalf("expected error %v, got %v", domain.ErrCollectStruct, err)
		}

		if html != "" {
			t.Fatalf("expected empty html, got %q", html)
		}
	})
}
