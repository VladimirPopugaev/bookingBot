package usecase

import (
	"booking_bot/internal/domain"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

type stubSiteWorkerRepository struct {
	fetchFn func(ctx context.Context, fetchURL string) (string, error)
	parseFn func(ctx context.Context, htmlReader io.Reader) (*domain.SiteInfo, error)

	fetchCalls     int
	parseCalls     int
	lastFetchURL   string
	lastParsedHTML string
}

func (s *stubSiteWorkerRepository) FetchSiteStruct(ctx context.Context, fetchURL string) (string, error) {
	s.fetchCalls++
	s.lastFetchURL = fetchURL

	if s.fetchFn == nil {
		return "", nil
	}

	return s.fetchFn(ctx, fetchURL)
}

func (s *stubSiteWorkerRepository) ParseSiteStruct(ctx context.Context, htmlReader io.Reader) (*domain.SiteInfo, error) {
	s.parseCalls++

	body, err := io.ReadAll(htmlReader)
	if err != nil {
		return nil, err
	}
	s.lastParsedHTML = string(body)

	if s.parseFn == nil {
		return nil, nil
	}

	return s.parseFn(ctx, strings.NewReader(s.lastParsedHTML))
}

func (s *stubSiteWorkerRepository) Close() error {
	return nil
}

func TestUsecase_AnalyzeSite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		rawURL         string
		repo           *stubSiteWorkerRepository
		wantErr        error
		wantInfo       *domain.SiteInfo
		wantFetchCalls int
		wantParseCalls int
		wantFetchURL   string
		wantParsedHTML string
	}{
		{
			name:           "empty url",
			rawURL:         "   ",
			repo:           &stubSiteWorkerRepository{},
			wantErr:        domain.ErrEmptyParameter,
			wantFetchCalls: 0,
			wantParseCalls: 0,
		},
		{
			name:           "invalid url",
			rawURL:         "not-a-url",
			repo:           &stubSiteWorkerRepository{},
			wantErr:        domain.ErrURLParse,
			wantFetchCalls: 0,
			wantParseCalls: 0,
		},
		{
			name:   "fetch site structure failed",
			rawURL: "https://example.com/path?q=1",
			repo: &stubSiteWorkerRepository{
				fetchFn: func(ctx context.Context, fetchURL string) (string, error) {
					return "", errors.New("upstream fetch failed")
				},
			},
			wantErr:        domain.ErrCollectStruct,
			wantFetchCalls: 1,
			wantParseCalls: 0,
			wantFetchURL:   "https://example.com/path?q=1",
		},
		{
			name:   "parse site structure failed",
			rawURL: " https://example.com/path?q=1 ",
			repo: &stubSiteWorkerRepository{
				fetchFn: func(ctx context.Context, fetchURL string) (string, error) {
					return "<html><body>payload</body></html>", nil
				},
				parseFn: func(ctx context.Context, htmlReader io.Reader) (*domain.SiteInfo, error) {
					return nil, errors.New("parse failed")
				},
			},
			wantErr:        domain.ErrParseStruct,
			wantFetchCalls: 1,
			wantParseCalls: 1,
			wantFetchURL:   "https://example.com/path?q=1",
			wantParsedHTML: "<html><body>payload</body></html>",
		},
		{
			name:   "success",
			rawURL: " https://example.com/path?q=1 ",
			repo: &stubSiteWorkerRepository{
				fetchFn: func(ctx context.Context, fetchURL string) (string, error) {
					return "<html><body>payload</body></html>", nil
				},
				parseFn: func(ctx context.Context, htmlReader io.Reader) (*domain.SiteInfo, error) {
					return &domain.SiteInfo{
						Title:       "Example",
						H1:          "Booking",
						LinksCount:  3,
						TextPreview: "payload preview",
					}, nil
				},
			},
			wantInfo: &domain.SiteInfo{
				Title:       "Example",
				H1:          "Booking",
				LinksCount:  3,
				TextPreview: "payload preview",
			},
			wantFetchCalls: 1,
			wantParseCalls: 1,
			wantFetchURL:   "https://example.com/path?q=1",
			wantParsedHTML: "<html><body>payload</body></html>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, err := New(nil, tt.repo, zerolog.Nop())
			if err != nil {
				t.Fatalf("expected no error creating usecase, got %v", err)
			}

			info, err := uc.AnalyzeSite(context.Background(), tt.rawURL)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if tt.wantErr != nil {
				if info != nil {
					t.Fatalf("expected nil info, got %+v", *info)
				}
			} else {
				if info == nil {
					t.Fatal("expected site info, got nil")
				}

				if *info != *tt.wantInfo {
					t.Fatalf("expected info %+v, got %+v", *tt.wantInfo, *info)
				}
			}

			if tt.repo.fetchCalls != tt.wantFetchCalls {
				t.Fatalf("expected fetch calls %d, got %d", tt.wantFetchCalls, tt.repo.fetchCalls)
			}

			if tt.repo.parseCalls != tt.wantParseCalls {
				t.Fatalf("expected parse calls %d, got %d", tt.wantParseCalls, tt.repo.parseCalls)
			}

			if tt.repo.lastFetchURL != tt.wantFetchURL {
				t.Fatalf("expected fetch URL %q, got %q", tt.wantFetchURL, tt.repo.lastFetchURL)
			}

			if tt.repo.lastParsedHTML != tt.wantParsedHTML {
				t.Fatalf("expected parsed HTML %q, got %q", tt.wantParsedHTML, tt.repo.lastParsedHTML)
			}
		})
	}
}
