package siteworker

import (
	"booking_bot/internal/domain"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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
				TargetURL:          "https://example.com",
				Timeout:            5 * time.Second,
				MonitoringInterval: 30 * time.Second,
				disableMonitoring:  true,
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

				if workerRepo.cfg.TargetURL != cfg.TargetURL {
					t.Fatalf("expected target url %q, got %q", cfg.TargetURL, workerRepo.cfg.TargetURL)
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
				TargetURL:          "https://example.com",
				Timeout:            5 * time.Second,
				MonitoringInterval: 0,
				disableMonitoring:  true,
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

				if cfg.MonitoringInterval != defaultMonitoringInterval {
					t.Fatalf(
						"expected input config monitoring interval to be normalized to %s, got %s",
						defaultMonitoringInterval,
						cfg.MonitoringInterval,
					)
				}
			},
		},
		{
			name: "empty target url",
			cfg: &Config{
				TargetURL:          "",
				Timeout:            5 * time.Second,
				MonitoringInterval: 30 * time.Second,
				disableMonitoring:  true,
			},
			wantErr: domain.ErrEmptyParameter,
		},
		{
			name: "blank target url",
			cfg: &Config{
				TargetURL:          "   ",
				Timeout:            5 * time.Second,
				MonitoringInterval: 30 * time.Second,
				disableMonitoring:  true,
			},
			wantErr: domain.ErrEmptyParameter,
		},
		{
			name: "invalid target url",
			cfg: &Config{
				TargetURL:          "://broken-url",
				Timeout:            5 * time.Second,
				MonitoringInterval: 30 * time.Second,
				disableMonitoring:  true,
			},
			wantErr: domain.ErrURLParse,
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
			TargetURL:          server.URL,
			Timeout:            5 * time.Second,
			MonitoringInterval: 30 * time.Second,
			disableMonitoring:  true,
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
			TargetURL:          serverURL,
			Timeout:            500 * time.Millisecond,
			MonitoringInterval: 30 * time.Second,
			disableMonitoring:  true,
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
			TargetURL:          server.URL,
			Timeout:            5 * time.Second,
			MonitoringInterval: 30 * time.Second,
			disableMonitoring:  true,
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
			TargetURL:          server.URL,
			Timeout:            50 * time.Millisecond,
			MonitoringInterval: 30 * time.Second,
			disableMonitoring:  true,
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

		html, err := workerRepo.FetchSiteStruct(context.Background())
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

	t.Run("starts immediately and repeats on interval", func(t *testing.T) {
		t.Parallel()

		var requests atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<html><body>ok</body></html>"))
		}))
		defer server.Close()

		repo, err := New(&Config{
			TargetURL:          server.URL,
			Timeout:            100 * time.Millisecond,
			MonitoringInterval: 25 * time.Millisecond,
		}, zerolog.Nop())
		if err != nil {
			t.Fatalf("expected no error creating repository, got %v", err)
		}
		t.Cleanup(func() {
			_ = repo.Close()
		})

		waitForRequests(t, &requests, 1, 300*time.Millisecond)
		waitForRequests(t, &requests, 2, 300*time.Millisecond)
	})

	t.Run("continues after fetch error", func(t *testing.T) {
		t.Parallel()

		var requests atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}))
		defer server.Close()

		repo, err := New(&Config{
			TargetURL:          server.URL,
			Timeout:            100 * time.Millisecond,
			MonitoringInterval: 25 * time.Millisecond,
		}, zerolog.Nop())
		if err != nil {
			t.Fatalf("expected no error creating repository, got %v", err)
		}
		t.Cleanup(func() {
			_ = repo.Close()
		})

		waitForRequests(t, &requests, 2, 400*time.Millisecond)
	})

	t.Run("close stops monitoring without hanging", func(t *testing.T) {
		t.Parallel()

		var requests atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<html><body>ok</body></html>"))
		}))
		defer server.Close()

		repo, err := New(&Config{
			TargetURL:          server.URL,
			Timeout:            100 * time.Millisecond,
			MonitoringInterval: 25 * time.Millisecond,
		}, zerolog.Nop())
		if err != nil {
			t.Fatalf("expected no error creating repository, got %v", err)
		}

		waitForRequests(t, &requests, 1, 300*time.Millisecond)

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

func waitForRequests(t *testing.T, requests *atomic.Int32, want int32, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if requests.Load() >= want {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("expected at least %d requests, got %d", want, requests.Load())
}
