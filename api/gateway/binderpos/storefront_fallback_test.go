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
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-direct"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dedicated"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-direct"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-dynamic"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dynamic"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "api-dedicated" {
			t.Fatalf("expected dedicated api card, got %+v", cards)
		}
	})

	t.Run("falls back to direct api before any dynamic attempt", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated api failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-direct"}}, nil },
			func() ([]gateway.Card, error) {
				t.Fatal("scrap dedicated should not run after direct api success")
				return nil, nil
			},
			func() ([]gateway.Card, error) {
				t.Fatal("scrap direct should not run after direct api success")
				return nil, nil
			},
			func() ([]gateway.Card, error) {
				t.Fatal("dynamic api should not run after direct api success")
				return nil, nil
			},
			func() ([]gateway.Card, error) {
				t.Fatal("dynamic scrape should not run after direct api success")
				return nil, nil
			},
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "api-direct" {
			t.Fatalf("expected direct api card, got %+v", cards)
		}
	})

	t.Run("falls back to dedicated scraper when api attempts fail", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("direct api failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dedicated"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-direct"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-dynamic"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dynamic"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-dedicated" {
			t.Fatalf("expected dedicated scraper card, got %+v", cards)
		}
	})

	t.Run("falls back to direct scraper before any dynamic attempt", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("direct api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated scraper failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-direct"}}, nil },
			func() ([]gateway.Card, error) {
				t.Fatal("dynamic api should not run after direct scrape success")
				return nil, nil
			},
			func() ([]gateway.Card, error) {
				t.Fatal("dynamic scrape should not run after direct scrape success")
				return nil, nil
			},
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-direct" {
			t.Fatalf("expected direct scraper card, got %+v", cards)
		}
	})

	t.Run("falls back to dynamic api only after cheaper attempts fail", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("direct api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated scraper failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("direct scraper failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "api-dynamic"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dynamic"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "api-dynamic" {
			t.Fatalf("expected dynamic api card, got %+v", cards)
		}
	})

	t.Run("falls back to dynamic scraper when every other attempt fails", func(t *testing.T) {
		cards, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("direct api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated scraper failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("direct scraper failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("dynamic api failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dynamic"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-dynamic" {
			t.Fatalf("expected dynamic scraper card, got %+v", cards)
		}
	})

	t.Run("returns final dynamic scraper error when all fail", func(t *testing.T) {
		apiDedicatedErr := errors.New("dedicated api failed")
		apiDirectErr := errors.New("direct api failed")
		scrapDedicatedErr := errors.New("dedicated scraper failed")
		directScrapErr := errors.New("direct scraper failed")
		apiDynamicErr := errors.New("dynamic api failed")
		scrapDynamicErr := errors.New("dynamic scraper failed")
		_, err := searchWithFallback(
			func() ([]gateway.Card, error) { return nil, apiDedicatedErr },
			func() ([]gateway.Card, error) { return nil, apiDirectErr },
			func() ([]gateway.Card, error) { return nil, scrapDedicatedErr },
			func() ([]gateway.Card, error) { return nil, directScrapErr },
			func() ([]gateway.Card, error) { return nil, apiDynamicErr },
			func() ([]gateway.Card, error) { return nil, scrapDynamicErr },
		)
		if !errors.Is(err, scrapDynamicErr) {
			t.Fatalf("expected dynamic scraper error, got %v", err)
		}
		expectedError := "attempt 6 (scrap-dynamic): dynamic scraper failed"
		if err == nil || err.Error() != expectedError {
			t.Fatalf("expected final error %q, got %v", expectedError, err)
		}
	})

	t.Run("runs attempts in the requested order", func(t *testing.T) {
		sequence := make([]string, 0, 6)
		fail := func(label string) func() ([]gateway.Card, error) {
			return func() ([]gateway.Card, error) {
				sequence = append(sequence, label)
				return nil, fmt.Errorf("%s failed", label)
			}
		}

		_, err := searchWithFallback(
			fail("api-dedicated"),
			fail("api-direct"),
			fail("scrap-dedicated"),
			fail("scrap-direct"),
			fail("api-dynamic"),
			fail("scrap-dynamic"),
		)
		if err == nil {
			t.Fatalf("expected fallback chain to return the final error")
		}
		expected := []string{"api-dedicated", "api-direct", "scrap-dedicated", "scrap-direct", "api-dynamic", "scrap-dynamic"}
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
			func() ([]gateway.Card, error) { return nil, errors.New("direct api failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated scraper failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("direct scraper failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("dynamic api failed") },
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

func TestSearchWithScrapDedicatedThenDirectThenDynamic(t *testing.T) {
	t.Run("returns dedicated scraper results on first success", func(t *testing.T) {
		cards, err := searchWithScrapDedicatedThenDirectThenDynamic(
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dedicated"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-direct"}}, nil },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dynamic"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-dedicated" {
			t.Fatalf("expected dedicated scraper card, got %+v", cards)
		}
	})

	t.Run("falls back to direct when dedicated fails", func(t *testing.T) {
		cards, err := searchWithScrapDedicatedThenDirectThenDynamic(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated scrap failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-direct"}}, nil },
			func() ([]gateway.Card, error) {
				t.Fatal("dynamic scrape should not run after direct success")
				return nil, nil
			},
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-direct" {
			t.Fatalf("expected direct scraper card, got %+v", cards)
		}
	})

	t.Run("falls back to dynamic when dedicated and direct fail", func(t *testing.T) {
		cards, err := searchWithScrapDedicatedThenDirectThenDynamic(
			func() ([]gateway.Card, error) { return nil, errors.New("dedicated scrap failed") },
			func() ([]gateway.Card, error) { return nil, errors.New("direct scrap failed") },
			func() ([]gateway.Card, error) { return []gateway.Card{{Name: "scrap-dynamic"}}, nil },
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-dynamic" {
			t.Fatalf("expected dynamic scraper card, got %+v", cards)
		}
	})

	t.Run("runs attempts in the requested order", func(t *testing.T) {
		sequence := make([]string, 0, 3)
		fail := func(label string) func() ([]gateway.Card, error) {
			return func() ([]gateway.Card, error) {
				sequence = append(sequence, label)
				return nil, fmt.Errorf("%s failed", label)
			}
		}

		_, err := searchWithScrapDedicatedThenDirectThenDynamic(
			fail("scrap-dedicated"),
			fail("scrap-direct"),
			fail("scrap-dynamic"),
		)
		if err == nil {
			t.Fatalf("expected fallback chain to return the final error")
		}
		expected := []string{"scrap-dedicated", "scrap-direct", "scrap-dynamic"}
		if len(sequence) != len(expected) {
			t.Fatalf("expected %d attempts, got %d (%v)", len(expected), len(sequence), sequence)
		}
		for i := range expected {
			if sequence[i] != expected[i] {
				t.Fatalf("attempt %d: expected %q, got %q", i+1, expected[i], sequence[i])
			}
		}
	})

	t.Run("returns final dynamic error when all fail", func(t *testing.T) {
		dedErr := errors.New("dedicated failed")
		dirErr := errors.New("direct failed")
		dynErr := errors.New("dynamic failed")
		_, err := searchWithScrapDedicatedThenDirectThenDynamic(
			func() ([]gateway.Card, error) { return nil, dedErr },
			func() ([]gateway.Card, error) { return nil, dirErr },
			func() ([]gateway.Card, error) { return nil, dynErr },
		)
		if !errors.Is(err, dynErr) {
			t.Fatalf("expected dynamic scraper error, got %v", err)
		}
		if err == nil || err.Error() != "attempt 3 (scrap-dynamic): dynamic failed" {
			t.Fatalf("unexpected final error: %v", err)
		}
	})
}
