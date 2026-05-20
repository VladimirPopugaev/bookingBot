package parser

import (
	"booking_bot/internal/domain"
	"context"
	"io"
	"strings"

	"github.com/rs/zerolog"
	"golang.org/x/net/html"
)

var (
	unavailableRegistrationPhrases = []string{
		"регистрация закрыта",
		"регистрация окончена",
		"регистрация завершена",
		"продажа завершена",
		"продажа билетов завершена",
		"билеты закончились",
		"нет билетов",
		"мест нет",
		"sold out",
		"registration closed",
		"registration ended",
		"tickets unavailable",
	}

	availableRegistrationPhrases = []string{
		"зарегистрироваться",
		"регистрация открыта",
		"купить билет",
		"купить билеты",
		"забронировать",
		"принять участие",
		"register",
		"registration is open",
		"buy ticket",
		"buy tickets",
		"book now",
	}
)

type repository struct {
	log zerolog.Logger
}

func New(logger zerolog.Logger) (domain.SiteParserRepository, error) {
	log := logger.With().Str("repository", "site_parser").Logger()

	return &repository{
		log: log,
	}, nil
}

func (r *repository) ParseSiteStruct(ctx context.Context, htmlReader io.Reader) (*domain.SiteInfo, error) {
	log := r.log.With().Str("method", "ParseSiteStruct").Logger()

	tokenizer := html.NewTokenizer(htmlReader)
	info := &domain.SiteInfo{
		IsRegistrationAvailable: false,
	}

	var (
		isInsideTitle bool
		isInsideH1    bool
	)
	pageTitle := ""

	for {
		select {
		case <-ctx.Done():
			log.Error().Err(ctx.Err()).Msg("Parse site struct cancelled")
			return nil, domain.ErrContextCancelled
		default:

		}

		switch tokenizer.Next() {
		case html.ErrorToken:
			err := tokenizer.Err()
			if err != nil && err != io.EOF {
				log.Error().Err(err).Msg("Parse site struct failed")
				return nil, err
			}

			info.Title = strings.TrimSpace(info.Title)
			if info.Title == "" {
				info.Title = strings.TrimSpace(pageTitle)
			}

			log.Trace().Msg("Parsed site struct successfully")
			return info, nil

		case html.StartTagToken:
			token := tokenizer.Token()

			switch token.DataAtom.String() {
			case "title":
				isInsideTitle = true
			case "h1":
				if info.Title == "" {
					isInsideH1 = true
				}
			}

		case html.EndTagToken:
			token := tokenizer.Token()

			switch token.DataAtom.String() {
			case "title":
				isInsideTitle = false
			case "h1":
				isInsideH1 = false
			}

		case html.TextToken:
			text := normalizeHTMLText(string(tokenizer.Text()))
			if text == "" {
				continue
			}

			if isInsideH1 {
				info.Title = appendSentencePart(info.Title, text)
			}

			if isInsideTitle {
				pageTitle = appendSentencePart(pageTitle, text)
			}
		}
	}
}

func (r *repository) IsAvailableToRegister(ctx context.Context, text string) (bool, error) {
	log := r.log.With().Str("method", "IsAvailableToRegister").Logger()

	visibleText, err := extractVisibleText(ctx, strings.NewReader(text), log)
	if err != nil {
		log.Error().Err(err).Msg("Extract visible text failed")
		return false, err
	}

	normalizedText := strings.ToLower(visibleText)
	if containsAnyPhrase(normalizedText, unavailableRegistrationPhrases) {
		return false, nil
	}

	return containsAnyPhrase(normalizedText, availableRegistrationPhrases), nil
}

func extractVisibleText(ctx context.Context, htmlReader io.Reader, log zerolog.Logger) (string, error) {
	tokenizer := html.NewTokenizer(htmlReader)
	builder := strings.Builder{}

	var depthLevel int

	for {
		select {
		case <-ctx.Done():
			log.Error().Err(ctx.Err()).Msg("Extract visible text cancelled")
			return "", domain.ErrContextCancelled
		default:
		}

		switch tokenizer.Next() {
		case html.ErrorToken:
			err := tokenizer.Err()
			if err != nil && err != io.EOF {
				log.Error().Err(err).Msg("Extract visible text failed")
				return "", err
			}

			return strings.TrimSpace(builder.String()), nil

		case html.StartTagToken:
			switch tokenizer.Token().DataAtom.String() {
			case "script", "style", "noscript":
				depthLevel++
			}

		case html.EndTagToken:
			switch tokenizer.Token().DataAtom.String() {
			case "script", "style", "noscript":
				if depthLevel > 0 {
					depthLevel--
				}
			}

		case html.TextToken:
			if depthLevel > 0 {
				continue
			}

			text := normalizeHTMLText(string(tokenizer.Text()))
			if text == "" {
				continue
			}

			if builder.Len() > 0 {
				builder.WriteByte(' ')
			}
			builder.WriteString(text)
		}
	}
}

func containsAnyPhrase(text string, phrases []string) bool {
	for _, phrase := range phrases {
		if strings.Contains(text, phrase) {
			return true
		}
	}

	return false
}

func normalizeHTMLText(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func appendSentencePart(current string, next string) string {
	if current == "" {
		return next
	}

	if next == "" {
		return current
	}

	return current + " " + next
}

func (r *repository) Close() error {
	r.log.Trace().Msg("Site parser repository closed")
	return nil
}
