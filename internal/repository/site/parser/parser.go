package parser

import (
	"booking_bot/internal/domain"
	"context"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/rs/zerolog"
	"golang.org/x/net/html"
)

var (
	defaultTextPreviewLength = 200
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
	info := &domain.SiteInfo{}

	var (
		isInsideTitle bool
		isInsideH1    bool
		depthLevel    int
	)

	previewBuilder := strings.Builder{}
	previewBuilder.Grow(defaultTextPreviewLength)

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
			info.H1 = strings.TrimSpace(info.H1)
			info.TextPreview = strings.TrimSpace(previewBuilder.String())

			log.Trace().Msg("Parsed site struct successfully")
			return info, nil

		case html.StartTagToken:
			token := tokenizer.Token()

			switch token.DataAtom.String() {
			case "title":
				isInsideTitle = true
			case "h1":
				if info.H1 == "" {
					isInsideH1 = true
				}
			case "a":
				info.LinksCount++
			case "script", "style", "noscript":
				depthLevel++
			}

		case html.EndTagToken:
			token := tokenizer.Token()

			switch token.DataAtom.String() {
			case "title":
				isInsideTitle = false
			case "h1":
				isInsideH1 = false
			case "script", "style", "noscript":
				if depthLevel > 0 {
					depthLevel--
				}
			}

		case html.SelfClosingTagToken:
			if tokenizer.Token().DataAtom.String() == "a" {
				info.LinksCount++
			}

		case html.TextToken:
			if depthLevel > 0 {
				continue
			}

			text := normalizeHTMLText(string(tokenizer.Text()))
			if text == "" {
				continue
			}

			if isInsideTitle {
				info.Title = appendSentencePart(info.Title, text)
			}

			if isInsideH1 {
				info.H1 = appendSentencePart(info.H1, text)
			}

			appendPreviewText(&previewBuilder, text, defaultTextPreviewLength)
		}
	}
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

func appendPreviewText(builder *strings.Builder, text string, limit int) {
	if utf8.RuneCountInString(builder.String()) >= limit || text == "" {
		return
	}

	remaining := limit - utf8.RuneCountInString(builder.String())
	if builder.Len() > 0 {
		if remaining == 0 {
			return
		}

		builder.WriteByte(' ')
		remaining--
	}

	if remaining <= 0 {
		return
	}

	if utf8.RuneCountInString(text) > remaining {
		builder.WriteString(truncateRunes(text, remaining))
		return
	}

	builder.WriteString(text)
}

func truncateRunes(text string, limit int) string {
	if limit <= 0 {
		return ""
	}

	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}

	return string(runes[:limit])
}

func (r *repository) Close() error {
	r.log.Trace().Msg("Site parser repository closed")
	return nil
}
