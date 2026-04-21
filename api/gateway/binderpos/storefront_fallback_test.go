package binderpos

import (
	"errors"
	"testing"
)

func TestSearchWithFallback(t *testing.T) {
	t.Run("returns dedicated api results first", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]cardLike, error) { return []cardLike{{Name: "api-dedicated"}}, nil },
			func() ([]cardLike, error) { return []cardLike{{Name: "scrap"}}, nil },
			func() ([]cardLike, error) { return []cardLike{{Name: "api-direct"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "api-dedicated" {
			t.Fatalf("expected dedicated api card, got %+v", cards)
		}
	})

	t.Run("falls back to scraper before direct api", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]cardLike, error) { return nil, errors.New("dedicated api failed") },
			func() ([]cardLike, error) { return []cardLike{{Name: "scrap"}}, nil },
			func() ([]cardLike, error) { return []cardLike{{Name: "api-direct"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap" {
			t.Fatalf("expected scraper card, got %+v", cards)
		}
	})

	t.Run("falls back to direct api when dedicated api and scraper fail", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]cardLike, error) { return nil, errors.New("dedicated api failed") },
			func() ([]cardLike, error) { return nil, errors.New("scraper failed") },
			func() ([]cardLike, error) { return []cardLike{{Name: "api-direct"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "api-direct" {
			t.Fatalf("expected direct api card, got %+v", cards)
		}
	})

	t.Run("returns final direct api error when all fail", func(t *testing.T) {
		dedicatedErr := errors.New("dedicated api failed")
		scrapErr := errors.New("scraper failed")
		directErr := errors.New("direct api failed")
		_, err := searchWithFallback(
			func() ([]cardLike, error) { return nil, dedicatedErr },
			func() ([]cardLike, error) { return nil, scrapErr },
			func() ([]cardLike, error) { return nil, directErr },
		)
		if !errors.Is(err, directErr) {
			t.Fatalf("expected direct api error, got %v", err)
		}
	})
}

type cardLike struct {
	Name string
}
