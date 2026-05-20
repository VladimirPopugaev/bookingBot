package http

import (
	"booking_bot/internal/domain"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

type stubUsecase struct {
	checkSiteForRegistrationFunc func(ctx context.Context, rawURL string) (*domain.SiteInfo, error)
}

func (s stubUsecase) CheckSiteForRegistration(ctx context.Context, rawURL string) (*domain.SiteInfo, error) {
	if s.checkSiteForRegistrationFunc == nil {
		return nil, nil
	}

	return s.checkSiteForRegistrationFunc(ctx, rawURL)
}

func (s stubUsecase) Close() {}

func TestHandler_GetSiteInfo(t *testing.T) {
	t.Parallel()

	logger := zerolog.Nop()

	t.Run("missing url returns bad request", func(t *testing.T) {
		t.Parallel()

		router := NewRouter(stubUsecase{
			checkSiteForRegistrationFunc: func(ctx context.Context, rawURL string) (*domain.SiteInfo, error) {
				t.Fatal("CheckSiteForRegistration must not be called when url is missing")
				return nil, nil
			},
		}, &logger)

		req := httptest.NewRequest(http.MethodGet, "/site-info", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("invalid url returns bad request", func(t *testing.T) {
		t.Parallel()

		router := NewRouter(stubUsecase{
			checkSiteForRegistrationFunc: func(ctx context.Context, rawURL string) (*domain.SiteInfo, error) {
				return nil, domain.ErrURLParse
			},
		}, &logger)

		req := httptest.NewRequest(http.MethodGet, "/site-info?url=bad-url", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("valid url returns mock response", func(t *testing.T) {
		t.Parallel()

		router := NewRouter(stubUsecase{
			checkSiteForRegistrationFunc: func(ctx context.Context, rawURL string) (*domain.SiteInfo, error) {
				return &domain.SiteInfo{
					URL:                     "https://example.com",
					Title:                   "Mock Title",
					IsRegistrationAvailable: true,
				}, nil
			},
		}, &logger)

		req := httptest.NewRequest(http.MethodGet, "/site-info?url=https://example.com", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response domain.SiteInfoResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if response.URL != "https://example.com" {
			t.Fatalf("expected url %q, got %q", "https://example.com", response.URL)
		}

		if response.Title != "Mock Title" {
			t.Fatalf("expected title %q, got %q", "Mock Title", response.Title)
		}

		if !response.IsRegistrationAvailable {
			t.Fatal("expected registration availability to be true")
		}
	})
}
