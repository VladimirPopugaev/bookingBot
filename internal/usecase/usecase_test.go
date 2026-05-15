package usecase

import (
	"booking_bot/internal/domain"
	"context"
	"errors"
	"testing"

	"github.com/rs/zerolog"
)

func TestUsecase_AnalyzeSite(t *testing.T) {
	t.Parallel()

	uc, err := New(nil, nil, zerolog.Nop())
	if err != nil {
		t.Fatalf("expected no error creating usecase, got %v", err)
	}

	tests := []struct {
		name      string
		rawURL    string
		wantErr   error
		wantTitle string
	}{
		{
			name:    "empty url",
			rawURL:  "",
			wantErr: domain.ErrEmptyParameter,
		},
		{
			name:    "invalid url",
			rawURL:  "not-a-url",
			wantErr: domain.ErrURLParse,
		},
		{
			name:      "valid url",
			rawURL:    "https://example.com",
			wantTitle: "Mock Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			info, err := uc.AnalyzeSite(context.Background(), tt.rawURL)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}

				if info != nil {
					t.Fatalf("expected nil info, got %+v", *info)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if info == nil {
				t.Fatal("expected site info, got nil")
			}

			if info.Title != tt.wantTitle {
				t.Fatalf("expected title %q, got %q", tt.wantTitle, info.Title)
			}

			if info.TextPreview != "Mock preview for https://example.com" {
				t.Fatalf("unexpected preview %q", info.TextPreview)
			}
		})
	}
}
