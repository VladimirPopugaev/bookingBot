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

	fetchCalls   int
	lastFetchURL string
}

func (s *stubSiteWorkerRepository) FetchSiteStruct(ctx context.Context, fetchURL string) (string, error) {
	s.fetchCalls++
	s.lastFetchURL = fetchURL

	if s.fetchFn == nil {
		return "", nil
	}

	return s.fetchFn(ctx, fetchURL)
}

func (s *stubSiteWorkerRepository) Close() error {
	return nil
}

type stubSiteParserRepository struct {
	parseFn func(ctx context.Context, htmlReader io.Reader) (*domain.SiteInfo, error)
	checkFn func(ctx context.Context, text string) (bool, error)

	parseCalls      int
	lastParsedHTML  string
	checkCalls      int
	lastCheckedHTML string
}

func (s *stubSiteParserRepository) ParseSiteStruct(ctx context.Context, htmlReader io.Reader) (*domain.SiteInfo, error) {
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

func (s *stubSiteParserRepository) IsAvailableToRegister(ctx context.Context, text string) (bool, error) {
	s.checkCalls++
	s.lastCheckedHTML = text

	if s.checkFn == nil {
		return false, nil
	}

	return s.checkFn(ctx, text)
}

func (s *stubSiteParserRepository) Close() error {
	return nil
}

func TestUsecase_CheckSiteForRegistration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		rawURL          string
		workerRepo      *stubSiteWorkerRepository
		parserRepo      *stubSiteParserRepository
		wantErr         error
		wantInfo        *domain.SiteInfo
		wantFetchCalls  int
		wantParseCalls  int
		wantCheckCalls  int
		wantFetchURL    string
		wantParsedHTML  string
		wantCheckedHTML string
	}{
		{
			name:           "empty url",
			rawURL:         "   ",
			workerRepo:     &stubSiteWorkerRepository{},
			parserRepo:     &stubSiteParserRepository{},
			wantErr:        domain.ErrEmptyParameter,
			wantFetchCalls: 0,
			wantParseCalls: 0,
			wantCheckCalls: 0,
		},
		{
			name:           "invalid url",
			rawURL:         "not-a-url",
			workerRepo:     &stubSiteWorkerRepository{},
			parserRepo:     &stubSiteParserRepository{},
			wantErr:        domain.ErrURLParse,
			wantFetchCalls: 0,
			wantParseCalls: 0,
			wantCheckCalls: 0,
		},
		{
			name:   "fetch site structure failed",
			rawURL: "https://example.com/path?q=1",
			workerRepo: &stubSiteWorkerRepository{
				fetchFn: func(ctx context.Context, fetchURL string) (string, error) {
					return "", errors.New("upstream fetch failed")
				},
			},
			parserRepo:     &stubSiteParserRepository{},
			wantErr:        domain.ErrCollectStruct,
			wantFetchCalls: 1,
			wantParseCalls: 0,
			wantCheckCalls: 0,
			wantFetchURL:   "https://example.com/path?q=1",
		},
		{
			name:   "parse site structure failed",
			rawURL: " https://example.com/path?q=1 ",
			workerRepo: &stubSiteWorkerRepository{
				fetchFn: func(ctx context.Context, fetchURL string) (string, error) {
					return "<html><body>payload</body></html>", nil
				},
			},
			parserRepo: &stubSiteParserRepository{
				parseFn: func(ctx context.Context, htmlReader io.Reader) (*domain.SiteInfo, error) {
					return nil, errors.New("parse failed")
				},
			},
			wantErr:        domain.ErrParseStruct,
			wantFetchCalls: 1,
			wantParseCalls: 1,
			wantCheckCalls: 0,
			wantFetchURL:   "https://example.com/path?q=1",
			wantParsedHTML: "<html><body>payload</body></html>",
		},
		{
			name:   "availability check failed",
			rawURL: " https://example.com/path?q=1 ",
			workerRepo: &stubSiteWorkerRepository{
				fetchFn: func(ctx context.Context, fetchURL string) (string, error) {
					return "<html><body>payload</body></html>", nil
				},
			},
			parserRepo: &stubSiteParserRepository{
				parseFn: func(ctx context.Context, htmlReader io.Reader) (*domain.SiteInfo, error) {
					return &domain.SiteInfo{
						Title: "Example",
					}, nil
				},
				checkFn: func(ctx context.Context, text string) (bool, error) {
					return false, errors.New("check failed")
				},
			},
			wantErr:         domain.ErrParseStruct,
			wantFetchCalls:  1,
			wantParseCalls:  1,
			wantCheckCalls:  1,
			wantFetchURL:    "https://example.com/path?q=1",
			wantParsedHTML:  "<html><body>payload</body></html>",
			wantCheckedHTML: "<html><body>payload</body></html>",
		},
		{
			name:   "success",
			rawURL: " https://example.com/path?q=1 ",
			workerRepo: &stubSiteWorkerRepository{
				fetchFn: func(ctx context.Context, fetchURL string) (string, error) {
					return "<html><body>payload</body></html>", nil
				},
			},
			parserRepo: &stubSiteParserRepository{
				parseFn: func(ctx context.Context, htmlReader io.Reader) (*domain.SiteInfo, error) {
					return &domain.SiteInfo{
						Title: "Example",
					}, nil
				},
				checkFn: func(ctx context.Context, text string) (bool, error) {
					return true, nil
				},
			},
			wantInfo: &domain.SiteInfo{
				Title:                   "Example",
				URL:                     "https://example.com/path?q=1",
				IsRegistrationAvailable: true,
			},
			wantFetchCalls:  1,
			wantParseCalls:  1,
			wantCheckCalls:  1,
			wantFetchURL:    "https://example.com/path?q=1",
			wantParsedHTML:  "<html><body>payload</body></html>",
			wantCheckedHTML: "<html><body>payload</body></html>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, err := New(nil, tt.workerRepo, tt.parserRepo, nil, zerolog.Nop())
			if err != nil {
				t.Fatalf("expected no error creating usecase, got %v", err)
			}

			info, err := uc.CheckSiteForRegistration(context.Background(), tt.rawURL)
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

			if tt.workerRepo.fetchCalls != tt.wantFetchCalls {
				t.Fatalf("expected fetch calls %d, got %d", tt.wantFetchCalls, tt.workerRepo.fetchCalls)
			}

			if tt.parserRepo.parseCalls != tt.wantParseCalls {
				t.Fatalf("expected parse calls %d, got %d", tt.wantParseCalls, tt.parserRepo.parseCalls)
			}

			if tt.parserRepo.checkCalls != tt.wantCheckCalls {
				t.Fatalf("expected availability check calls %d, got %d", tt.wantCheckCalls, tt.parserRepo.checkCalls)
			}

			if tt.workerRepo.lastFetchURL != tt.wantFetchURL {
				t.Fatalf("expected fetch URL %q, got %q", tt.wantFetchURL, tt.workerRepo.lastFetchURL)
			}

			if tt.parserRepo.lastParsedHTML != tt.wantParsedHTML {
				t.Fatalf("expected parsed HTML %q, got %q", tt.wantParsedHTML, tt.parserRepo.lastParsedHTML)
			}

			if tt.parserRepo.lastCheckedHTML != tt.wantCheckedHTML {
				t.Fatalf("expected checked HTML %q, got %q", tt.wantCheckedHTML, tt.parserRepo.lastCheckedHTML)
			}
		})
	}
}
