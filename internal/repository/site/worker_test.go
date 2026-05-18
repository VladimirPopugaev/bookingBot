package site

import (
	"booking_bot/internal/domain"
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
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
				Timeout:            5 * time.Second,
				MonitoringInterval: 30 * time.Second,
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

				if workerRepo.cfg != cfg {
					t.Fatal("expected config to be same pointer, got the the copy")
				}

				if workerRepo.cfg.Timeout != cfg.Timeout {
					t.Fatalf("expected timeout %s, got %s", cfg.Timeout, workerRepo.cfg.Timeout)
				}

				if workerRepo.cfg.MonitoringInterval != cfg.MonitoringInterval {
					t.Fatalf(
						"expected monitoring interval %s, got %s",
						cfg.MonitoringInterval,
						workerRepo.cfg.MonitoringInterval,
					)
				}
			},
		},
		{
			name: "default monitoring interval applied",
			cfg: &Config{
				Timeout:            5 * time.Second,
				MonitoringInterval: 0,
			},
			assertNew: func(t *testing.T, repo any, cfg *Config) {
				t.Helper()

				workerRepo, ok := repo.(*worker)
				if !ok {
					t.Fatalf("expected repository type *worker, got %T", repo)
				}

				if workerRepo.cfg.MonitoringInterval != defaultMonitoringInterval {
					t.Fatalf(
						"expected default monitoring interval %s, got %s",
						defaultMonitoringInterval,
						workerRepo.cfg.MonitoringInterval,
					)
				}
			},
		},
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: nil,
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

			if tt.cfg == nil {
				if err == nil {
					t.Fatal("expected error, got nil")
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

			t.Cleanup(func() {
				_ = repo.Close()
			})

			if tt.assertNew != nil {
				tt.assertNew(t, repo, tt.cfg)
			}
		})
	}
}

func TestWorker_FetchSiteStruct(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping network integration test on Windows due to unstable colly/net-http interaction on Go 1.26")
	}

	t.Run("success", func(t *testing.T) {
		const expectedHTML = "<html><body>ok</body></html>"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(expectedHTML))
		}))
		defer server.Close()

		repo, err := New(&Config{
			Timeout:            5 * time.Second,
			MonitoringInterval: 30 * time.Second,
		}, zerolog.Nop())
		if err != nil {
			t.Fatalf("expected no error creating repository, got %v", err)
		}
		t.Cleanup(func() {
			_ = repo.Close()
		})

		workerRepo, ok := repo.(*worker)
		if !ok {
			t.Fatalf("expected repository type *worker, got %T", repo)
		}

		html, err := workerRepo.FetchSiteStruct(context.Background(), server.URL)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if html != expectedHTML {
			t.Fatalf("expected html %q, got %q", expectedHTML, html)
		}
	})

	t.Run("server unavailable", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("expected to reserve local port, got %v", err)
		}

		serverURL := "http://" + listener.Addr().String()
		_ = listener.Close()

		repo, err := New(&Config{
			Timeout:            500 * time.Millisecond,
			MonitoringInterval: 30 * time.Second,
		}, zerolog.Nop())
		if err != nil {
			t.Fatalf("expected no error creating repository, got %v", err)
		}
		t.Cleanup(func() {
			_ = repo.Close()
		})

		workerRepo, ok := repo.(*worker)
		if !ok {
			t.Fatalf("expected repository type *worker, got %T", repo)
		}

		html, err := workerRepo.FetchSiteStruct(context.Background(), serverURL)
		if !errors.Is(err, domain.ErrCollectStruct) {
			t.Fatalf("expected error %v, got %v", domain.ErrCollectStruct, err)
		}

		if html != "" {
			t.Fatalf("expected empty html, got %q", html)
		}
	})

	t.Run("server returns 500", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}))
		defer server.Close()

		repo, err := New(&Config{
			Timeout:            5 * time.Second,
			MonitoringInterval: 30 * time.Second,
		}, zerolog.Nop())
		if err != nil {
			t.Fatalf("expected no error creating repository, got %v", err)
		}
		t.Cleanup(func() {
			_ = repo.Close()
		})

		workerRepo, ok := repo.(*worker)
		if !ok {
			t.Fatalf("expected repository type *worker, got %T", repo)
		}

		html, err := workerRepo.FetchSiteStruct(context.Background(), server.URL)
		if !errors.Is(err, domain.ErrCollectStruct) {
			t.Fatalf("expected error %v, got %v", domain.ErrCollectStruct, err)
		}

		if html != "" {
			t.Fatalf("expected empty html, got %q", html)
		}
	})

	t.Run("request timeout exceeded", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<html><body>slow</body></html>"))
		}))
		defer server.Close()

		repo, err := New(&Config{
			Timeout:            50 * time.Millisecond,
			MonitoringInterval: 30 * time.Second,
		}, zerolog.Nop())
		if err != nil {
			t.Fatalf("expected no error creating repository, got %v", err)
		}
		t.Cleanup(func() {
			_ = repo.Close()
		})

		workerRepo, ok := repo.(*worker)
		if !ok {
			t.Fatalf("expected repository type *worker, got %T", repo)
		}

		html, err := workerRepo.FetchSiteStruct(context.Background(), server.URL)
		if !errors.Is(err, domain.ErrCollectStruct) {
			t.Fatalf("expected error %v, got %v", domain.ErrCollectStruct, err)
		}

		if html != "" {
			t.Fatalf("expected empty html, got %q", html)
		}
	})
}

func TestWorker_MonitoringLoop(t *testing.T) {
	t.Parallel()

	t.Run("repository does not start monitoring on its own", func(t *testing.T) {
		t.Parallel()

		repo, err := New(&Config{
			Timeout:            100 * time.Millisecond,
			MonitoringInterval: 25 * time.Millisecond,
		}, zerolog.Nop())
		if err != nil {
			t.Fatalf("expected no error creating repository, got %v", err)
		}
		t.Cleanup(func() {
			_ = repo.Close()
		})

		time.Sleep(150 * time.Millisecond)
	})

	t.Run("close without monitoring does not hang", func(t *testing.T) {
		t.Parallel()

		repo, err := New(&Config{
			Timeout:            100 * time.Millisecond,
			MonitoringInterval: 25 * time.Millisecond,
		}, zerolog.Nop())
		if err != nil {
			t.Fatalf("expected no error creating repository, got %v", err)
		}

		done := make(chan struct{})
		go func() {
			defer close(done)
			_ = repo.Close()
		}()

		select {
		case <-done:
		case <-time.After(300 * time.Millisecond):
			t.Fatal("expected Close to finish without hanging")
		}
	})
}
