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
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-shared"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-shared"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "api-dedicated" {
			t.Fatalf("expected dedicated api card, got %+v", cards)
		}
	})

	t.Run("falls back to shared api before dedicated scraper", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated api failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-shared"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-shared"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "api-shared" {
			t.Fatalf("expected shared api card, got %+v", cards)
		}
	})

	t.Run("falls back to dedicated scraper before shared-proxy scraper", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("shared api failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-shared"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap" {
			t.Fatalf("expected scraper card, got %+v", cards)
		}
	})

	t.Run("falls back to shared-proxy scraper when dedicated api, shared api and dedicated scraper fail", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("shared api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("scraper failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-shared"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-shared" {
			t.Fatalf("expected shared-proxy scraper card, got %+v", cards)
		}
	})

	t.Run("returns final shared-proxy scraper error when all fail", func(t *testing.T) {
		dedicatedErr := errors.New("dedicated api failed")
		sharedErr := errors.New("shared api failed")
		scrapErr := errors.New("scraper failed")
		sharedScrapErr := errors.New("shared scraper failed")
		_, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, dedicatedErr },
			func() ([]gateway.Card, error) { return nil, sharedErr },
			func() ([]gateway.Card, error) { return nil, scrapErr },
			func() ([]gateway.Card, error) { return nil, sharedScrapErr },
		)
		if !errors.Is(err, sharedScrapErr) {
			t.Fatalf("expected shared-proxy scraper error, got %v", err)
		}
		expectedError := "attempt 4 (scrap-shared): shared scraper failed"
		if err == nil || err.Error() != expectedError {
			t.Fatalf("expected final error %q, got %v", expectedError, err)
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
			fail("api-shared"),
			fail("scrap-dedicated"),
			fail("scrap-shared"),
		)
		if err == nil {
			t.Fatalf("expected fallback chain to return the final error")
		}
		expected := []string{"api-dedicated", "api-shared", "scrap-dedicated", "scrap-shared"}
		if len(sequence) != len(expected) {
			t.Fatalf("expected %d attempts, got %d (%v)", len(expected), len(sequence), sequence)
		}
		for i := range expected {
			if sequence[i] != expected[i] {
				t.Fatalf("attempt %d: expected %q, got %q", i+1, expected[i], sequence[i])
			}
		}
	})

	t.Run("returns no error when a fallback attempt succeeds with empty cards", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("shared api failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{}, nil },
			func() ([]gateway.Card, error) { return nil, errors.New("shared scraper failed") },
		)
		if err != nil {
			t.Fatalf("expected nil error when any attempt succeeds with empty result, got %v", err)
		}
		if len(cards) != 0 {
			t.Fatalf("expected zero cards, got %+v", cards)
		}
	})
}
