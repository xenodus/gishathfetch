package binderpos

import (
	"errors"
	"fmt"
	"testing"

	"mtg-price-checker-sg/gateway"
)

func TestRunFallbackAttempts(t *testing.T) {
	t.Run("returns the first attempt that succeeds without running later ones", func(t *testing.T) {
		cards, err := runFallbackAttempts(
			fallbackAttempt{strategy: "decklist-dedicated", fn: func() ([]gateway.Card, error) {
				return []gateway.Card{{Name: "decklist-dedicated"}}, nil
			}},
			fallbackAttempt{strategy: "decklist-direct", fn: func() ([]gateway.Card, error) {
				t.Fatal("later attempt should not run after the first succeeds")
				return nil, nil
			}},
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "decklist-dedicated" {
			t.Fatalf("expected first attempt card, got %+v", cards)
		}
	})

	t.Run("falls back to the next attempt only after an error", func(t *testing.T) {
		cards, err := runFallbackAttempts(
			fallbackAttempt{strategy: "decklist-dedicated", fn: func() ([]gateway.Card, error) {
				return nil, errors.New("decklist dedicated failed")
			}},
			fallbackAttempt{strategy: "scrap-dedicated", fn: func() ([]gateway.Card, error) {
				return []gateway.Card{{Name: "scrap-dedicated"}}, nil
			}},
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-dedicated" {
			t.Fatalf("expected fallback card, got %+v", cards)
		}
	})

	t.Run("continues after an empty but error-free attempt", func(t *testing.T) {
		cards, err := runFallbackAttempts(
			fallbackAttempt{strategy: "decklist-dedicated", fn: func() ([]gateway.Card, error) {
				return nil, errors.New("decklist dedicated failed")
			}},
			fallbackAttempt{strategy: "scrap-dedicated", fn: func() ([]gateway.Card, error) {
				return []gateway.Card{}, nil
			}},
			fallbackAttempt{strategy: "scrap-direct", fn: func() ([]gateway.Card, error) {
				return []gateway.Card{{Name: "scrap-direct"}}, nil
			}},
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-direct" {
			t.Fatalf("expected scrap-direct card, got %+v", cards)
		}
	})

	t.Run("returns empty only after the final attempt succeeds without cards", func(t *testing.T) {
		cards, err := runFallbackAttempts(
			fallbackAttempt{strategy: "decklist-dedicated", fn: func() ([]gateway.Card, error) {
				return []gateway.Card{}, nil
			}},
			fallbackAttempt{strategy: "scrap-direct", fn: func() ([]gateway.Card, error) {
				return []gateway.Card{}, nil
			}},
		)
		if err != nil {
			t.Fatalf("expected nil error when the final attempt is empty, got %v", err)
		}
		if len(cards) != 0 {
			t.Fatalf("expected zero cards, got %+v", cards)
		}
	})

	t.Run("runs attempts in order and returns the final annotated error", func(t *testing.T) {
		sequence := make([]string, 0, 6)
		fail := func(label string) fallbackAttempt {
			return fallbackAttempt{strategy: label, fn: func() ([]gateway.Card, error) {
				sequence = append(sequence, label)
				return nil, fmt.Errorf("%s failed", label)
			}}
		}

		lastErr := errors.New("scrap-dynamic failed")
		_, err := runFallbackAttempts(
			fail("decklist-dedicated"),
			fail("decklist-direct"),
			fail("scrap-dedicated"),
			fail("scrap-direct"),
			fail("decklist-dynamic"),
			fallbackAttempt{strategy: "scrap-dynamic", fn: func() ([]gateway.Card, error) {
				sequence = append(sequence, "scrap-dynamic")
				return nil, lastErr
			}},
		)
		if !errors.Is(err, lastErr) {
			t.Fatalf("expected wrapped final error, got %v", err)
		}
		expectedError := "attempt 6 (scrap-dynamic): scrap-dynamic failed"
		if err == nil || err.Error() != expectedError {
			t.Fatalf("expected final error %q, got %v", expectedError, err)
		}

		expected := []string{
			"decklist-dedicated", "decklist-direct",
			"scrap-dedicated", "scrap-direct",
			"decklist-dynamic", "scrap-dynamic",
		}
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
