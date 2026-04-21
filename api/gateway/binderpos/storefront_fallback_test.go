package binderpos

import (
	"errors"
	"fmt"
	"testing"

	"mtg-price-checker-sg/gateway"
)

func TestSearchWithFallback(t *testing.T) {
	t.Run("returns dedicated api results first", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-dedicated"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-shared"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-direct"}}, nil },
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
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated api failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-shared"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-direct"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap" {
			t.Fatalf("expected scraper card, got %+v", cards)
		}
	})

	t.Run("falls back to shared api before direct api", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("scraper failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-shared"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-direct"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "api-shared" {
			t.Fatalf("expected shared api card, got %+v", cards)
		}
	})

	t.Run("falls back to direct api when dedicated api, scraper and shared api fail", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("scraper failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("shared api failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-direct"}}, nil },
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
		sharedErr := errors.New("shared api failed")
		directErr := errors.New("direct api failed")
		_, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, dedicatedErr },
			func() ([]gateway.Card, error) { return nil, scrapErr },
			func() ([]gateway.Card, error) { return nil, sharedErr },
			func() ([]gateway.Card, error) { return nil, directErr },
		)
		if !errors.Is(err, directErr) {
			t.Fatalf("expected direct api error, got %v", err)
		}
	})

	t.Run("runs attempts in the requested order", func(t *testing.T) {
		sequence := make([]string, 0, 4)
		fail := func(label string) func() ([]gateway.Card, error) {
			return func() ([]gateway.Card, error) {
				sequence = append(sequence, label)
				return nil, fmt.Errorf("%s failed", label)
			}
		}

		_, err := searchWithFallback(
			fail("api-dedicated"),
			fail("scrap-dedicated"),
			fail("api-shared"),
			fail("direct"),
		)
		if err == nil {
			t.Fatalf("expected fallback chain to return the final error")
		}
		expected := []string{"api-dedicated", "scrap-dedicated", "api-shared", "direct"}
		if len(sequence) != len(expected) {
			t.Fatalf("expected %d attempts, got %d (%v)", len(expected), len(sequence), sequence)
		}
		for i := range expected {
			if sequence[i] != expected[i] {
				t.Fatalf("attempt %d: expected %q, got %q", i+1, expected[i], sequence[i])
			}
		}
	})
}
