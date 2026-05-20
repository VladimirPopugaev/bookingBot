package parser

import (
	"booking_bot/internal/domain"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestRepository_ParseSiteStruct(t *testing.T) {
	t.Parallel()

	newParser := func(t *testing.T) *repository {
		t.Helper()

		repo, err := New(zerolog.Nop())
		if err != nil {
			t.Fatalf("expected no error creating repository, got %v", err)
		}

		t.Cleanup(func() {
			_ = repo.Close()
		})

		parserRepo, ok := repo.(*repository)
		if !ok {
			t.Fatalf("expected repository type *repository, got %T", repo)
		}

		return parserRepo
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		html := `
<!doctype html>
<html>
	<head>
		<title> Booking   Page </title>
		<style>.hidden { display:none; }</style>
	</head>
	<body>
		<h1>Main Booking</h1>
		<p>Find the best option for your stay.</p>
		<a href="/one">One</a>
		<div>
			<a href="/two">Two</a>
		</div>
		<script>window.__SECRET__ = "skip me"</script>
	</body>
</html>`

		info, err := newParser(t).ParseSiteStruct(context.Background(), strings.NewReader(html))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if info.Title != "Main Booking" {
			t.Fatalf("expected title %q, got %q", "Main Booking", info.Title)
		}

		if info.URL != "" {
			t.Fatalf("expected empty url, got %q", info.URL)
		}

		if info.IsRegistrationAvailable {
			t.Fatal("expected registration availability to be false")
		}
	})

	t.Run("empty html", func(t *testing.T) {
		t.Parallel()

		info, err := newParser(t).ParseSiteStruct(context.Background(), strings.NewReader(""))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if info == nil {
			t.Fatal("expected site info, got nil")
		}

		if *info != (domain.SiteInfo{}) {
			t.Fatalf("expected empty site info, got %+v", *info)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		info, err := newParser(t).ParseSiteStruct(ctx, strings.NewReader("<html><body>test</body></html>"))
		if !errors.Is(err, domain.ErrContextCancelled) {
			t.Fatalf("expected error %v, got %v", domain.ErrContextCancelled, err)
		}

		if info != nil {
			t.Fatalf("expected nil site info, got %+v", *info)
		}
	})

	t.Run("falls back to title when h1 is missing", func(t *testing.T) {
		t.Parallel()

		html := "<html><head><title>Big Page</title></head><body><p>Body text</p></body></html>"

		info, err := newParser(t).ParseSiteStruct(context.Background(), strings.NewReader(html))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if info.Title != "Big Page" {
			t.Fatalf("expected title %q, got %q", "Big Page", info.Title)
		}

		if info.IsRegistrationAvailable {
			t.Fatal("expected registration availability to be false")
		}
	})
}

func TestRepository_IsAvailableToRegister(t *testing.T) {
	t.Parallel()

	newParser := func(t *testing.T) *repository {
		t.Helper()

		repo, err := New(zerolog.Nop())
		if err != nil {
			t.Fatalf("expected no error creating repository, got %v", err)
		}

		t.Cleanup(func() {
			_ = repo.Close()
		})

		parserRepo, ok := repo.(*repository)
		if !ok {
			t.Fatalf("expected repository type *repository, got %T", repo)
		}

		return parserRepo
	}

	tests := []struct {
		name      string
		text      string
		ctx       func() context.Context
		want      bool
		wantError error
	}{
		{
			name: "html with registration button is available",
			text: `<html><body><button>Зарегистрироваться</button></body></html>`,
			want: true,
		},
		{
			name: "html with buy ticket text is available",
			text: `<html><body><a href="/tickets">Купить билет</a></body></html>`,
			want: true,
		},
		{
			name: "closed registration is unavailable",
			text: `<html><body><p>Регистрация закрыта</p></body></html>`,
		},
		{
			name: "sold out tickets are unavailable",
			text: `<html><body><p>Билеты закончились</p></body></html>`,
		},
		{
			name: "unavailable phrase has priority",
			text: `<html><body><button>Зарегистрироваться</button><p>Регистрация окончена</p></body></html>`,
		},
		{
			name: "service tags are ignored",
			text: `<html><head><style>.button:after{content:"Купить билет"}</style></head><body><p>Описание события</p><script>register()</script><noscript>Зарегистрироваться</noscript></body></html>`,
		},
		{
			name: "plain text is supported",
			text: `Registration is open for this event`,
			want: true,
		},
		{
			name: "empty text is unavailable",
			text: "",
		},
		{
			name: "context cancelled",
			text: `<html><body>Зарегистрироваться</body></html>`,
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			wantError: domain.ErrContextCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			if tt.ctx != nil {
				ctx = tt.ctx()
			}

			got, err := newParser(t).IsAvailableToRegister(ctx, tt.text)
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("expected error %v, got %v", tt.wantError, err)
			}

			if got != tt.want {
				t.Fatalf("expected availability %t, got %t", tt.want, got)
			}
		})
	}
}
