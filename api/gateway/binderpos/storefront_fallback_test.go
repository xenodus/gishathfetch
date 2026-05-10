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
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-dynamic"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dedicated"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dynamic"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-direct"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "api-dedicated" {
			t.Fatalf("expected dedicated api card, got %+v", cards)
		}
	})

	t.Run("falls back to dynamic api when dedicated api fails", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated api failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-dynamic"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dedicated"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dynamic"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-direct"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "api-dynamic" {
			t.Fatalf("expected dynamic api card, got %+v", cards)
		}
	})

	t.Run("falls back to dedicated scraper when api attempts fail", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("dynamic api failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dedicated"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dynamic"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-direct"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-dedicated" {
			t.Fatalf("expected dedicated scraper card, got %+v", cards)
		}
	})

	t.Run("falls back to dynamic scraper when proxy api and dedicated scrape fail", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("dynamic api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated scraper failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dynamic"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-direct"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-dynamic" {
			t.Fatalf("expected dynamic scraper card, got %+v", cards)
		}
	})

	t.Run("falls back to direct scraper when all proxy attempts fail", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("dynamic api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated scraper failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("dynamic scraper failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-direct"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-direct" {
			t.Fatalf("expected direct scraper card, got %+v", cards)
		}
	})

	t.Run("returns final direct scraper error when all fail", func(t *testing.T) {
		apiDedicatedErr := errors.New("dedicated api failed")
		apiDynamicErr := errors.New("dynamic api failed")
		scrapDedicatedErr := errors.New("dedicated scraper failed")
		scrapDynamicErr := errors.New("dynamic scraper failed")
		directScrapErr := errors.New("direct scraper failed")
		_, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, apiDedicatedErr },
			func() ([]gateway.Card, error) { return nil, apiDynamicErr },
			func() ([]gateway.Card, error) { return nil, scrapDedicatedErr },
			func() ([]gateway.Card, error) { return nil, scrapDynamicErr },
			func() ([]gateway.Card, error) { return nil, directScrapErr },
		)
		if !errors.Is(err, directScrapErr) {
			t.Fatalf("expected direct scraper error, got %v", err)
		}
		expectedError := "attempt 5 (scrap-direct): direct scraper failed"
		if err == nil || err.Error() != expectedError {
			t.Fatalf("expected final error %q, got %v", expectedError, err)
		}
	})

	t.Run("runs attempts in the requested order", func(t *testing.T) {
		sequence := make([]string, 0, 5)
		fail := func(label string) func() ([]gateway.Card, error) {
			return func() ([]gateway.Card, error) {
				sequence = append(sequence, label)
				return nil, fmt.Errorf("%s failed", label)
			}
		}

		_, err := searchWithFallback(
			fail("api-dedicated"),
			fail("api-dynamic"),
			fail("scrap-dedicated"),
			fail("scrap-dynamic"),
			fail("scrap-direct"),
		)
		if err == nil {
			t.Fatalf("expected fallback chain to return the final error")
		}
		expected := []string{"api-dedicated", "api-dynamic", "scrap-dedicated", "scrap-dynamic", "scrap-direct"}
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
			func() ([]gateway.Card, error) { return nil, errors.New("dynamic api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated scraper failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("dynamic scraper failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error when any attempt succeeds with empty result, got %v", err)
		}
		if len(cards) != 0 {
			t.Fatalf("expected zero cards, got %+v", cards)
		}
	})
}

func TestSearchWithScrapDedicatedThenDynamicThenDirect(t *testing.T) {
	t.Run("returns dedicated scraper results on first success", func(t *testing.T) {
		cards, err := searchWithScrapDedicatedThenDynamicThenDirect(
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dedicated"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dynamic"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-direct"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-dedicated" {
			t.Fatalf("expected dedicated scraper card, got %+v", cards)
		}
	})

	t.Run("falls back to dynamic when dedicated fails", func(t *testing.T) {
		cards, err := searchWithScrapDedicatedThenDynamicThenDirect(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated scrap failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dynamic"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-direct"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-dynamic" {
			t.Fatalf("expected dynamic scraper card, got %+v", cards)
		}
	})

	t.Run("falls back to direct when dedicated and dynamic fail", func(t *testing.T) {
		cards, err := searchWithScrapDedicatedThenDynamicThenDirect(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated scrap failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("dynamic scrap failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-direct"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-direct" {
			t.Fatalf("expected direct scraper card, got %+v", cards)
		}
	})

	t.Run("returns final direct error when both fail", func(t *testing.T) {
		dedErr := errors.New("dedicated failed")
		dynErr := errors.New("dynamic failed")
		dirErr := errors.New("direct failed")
		_, err := searchWithScrapDedicatedThenDynamicThenDirect(
			func() ([]gateway.Card, error) { return nil, dedErr },
			func() ([]gateway.Card, error) { return nil, dynErr },
			func() ([]gateway.Card, error) { return nil, dirErr },
		)
		if !errors.Is(err, dirErr) {
			t.Fatalf("expected direct scraper error, got %v", err)
		}
		if err == nil || err.Error() != "attempt 3 (scrap-direct): direct failed" {
			t.Fatalf("unexpected final error: %v", err)
		}
	})
}
