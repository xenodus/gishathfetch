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
			fallbackAttempt{strategy: "scrap-dedicated", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
				return []gateway.Card{{Name: "scrap-dedicated"}}, nil
			}},
			fallbackAttempt{strategy: "scrap-direct", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
				t.Fatal("later attempt should not run after the first succeeds")
				return nil, nil
			}},
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-dedicated" {
			t.Fatalf("expected first attempt card, got %+v", cards)
		}
	})

	t.Run("falls back to the next attempt only after a non-5xx error", func(t *testing.T) {
		cards, err := runFallbackAttempts(
			fallbackAttempt{strategy: "scrap-dedicated", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
				return nil, errors.New("scrap dedicated failed")
			}},
			fallbackAttempt{strategy: "scrap-direct", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
				return []gateway.Card{{Name: "scrap-direct"}}, nil
			}},
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-direct" {
			t.Fatalf("expected fallback card, got %+v", cards)
		}
	})

	t.Run("falls back to decklist after scrap 429 errors", func(t *testing.T) {
		sequence := make([]string, 0, 3)
		cards, err := runFallbackAttempts(
			fallbackAttempt{strategy: "scrap-dedicated", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
				sequence = append(sequence, "scrap-dedicated")
				return nil, errors.New("unexpected status 429")
			}},
			fallbackAttempt{strategy: "scrap-direct", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
				sequence = append(sequence, "scrap-direct")
				return nil, errors.New("unexpected status 429")
			}},
			fallbackAttempt{strategy: "decklist-dedicated", family: strategyFamilyDecklist, fn: func() ([]gateway.Card, error) {
				sequence = append(sequence, "decklist-dedicated")
				return []gateway.Card{{Name: "decklist-dedicated"}}, nil
			}},
			fallbackAttempt{strategy: "scrap-dynamic", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
				t.Fatal("scrap-dynamic should not run before decklist after scrap 429 errors")
				return nil, nil
			}},
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "decklist-dedicated" {
			t.Fatalf("expected decklist fallback card, got %+v", cards)
		}
		want := []string{"scrap-dedicated", "scrap-direct", "decklist-dedicated"}
		if len(sequence) != len(want) {
			t.Fatalf("expected %v, got %v", want, sequence)
		}
		for i := range want {
			if sequence[i] != want[i] {
				t.Fatalf("attempt %d: expected %q, got %q", i+1, want[i], sequence[i])
			}
		}
	})

	t.Run("does not fall back after a scrap 5xx error", func(t *testing.T) {
		sequence := make([]string, 0, 2)
		_, err := runFallbackAttempts(
			fallbackAttempt{strategy: "scrap-dedicated", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
				sequence = append(sequence, "scrap-dedicated")
				return nil, errors.New("503 Service Unavailable")
			}},
			fallbackAttempt{strategy: "decklist-dedicated", family: strategyFamilyDecklist, fn: func() ([]gateway.Card, error) {
				t.Fatal("decklist should not run after a scrap 5xx error")
				return nil, nil
			}},
		)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !gateway.IsHTTPServerError(err) {
			t.Fatalf("expected 5xx error, got %v", err)
		}
		if len(sequence) != 1 || sequence[0] != "scrap-dedicated" {
			t.Fatalf("expected only scrap-dedicated, got %v", sequence)
		}
	})

	t.Run("when scrap is empty, returns final empty result without trying decklist", func(t *testing.T) {
		sequence := make([]string, 0, 2)
		cards, err := runFallbackAttempts(
			fallbackAttempt{strategy: "scrap-dedicated", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
				sequence = append(sequence, "scrap-dedicated")
				return []gateway.Card{}, nil
			}},
			fallbackAttempt{strategy: "decklist-dedicated", family: strategyFamilyDecklist, fn: func() ([]gateway.Card, error) {
				t.Fatal("decklist should not run after empty scrap-dedicated")
				return nil, nil
			}},
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 0 {
			t.Fatalf("expected zero cards, got %+v", cards)
		}
		if len(sequence) != 1 || sequence[0] != "scrap-dedicated" {
			t.Fatalf("expected only scrap-dedicated, got %v", sequence)
		}
	})

	t.Run("returns final empty result without trying later strategies after empty decklist", func(t *testing.T) {
		sequence := make([]string, 0, 3)
		cards, err := runFallbackAttempts(
			fallbackAttempt{strategy: "decklist-dedicated", family: strategyFamilyDecklist, fn: func() ([]gateway.Card, error) {
				sequence = append(sequence, "decklist-dedicated")
				return []gateway.Card{}, nil
			}},
			fallbackAttempt{strategy: "decklist-direct", family: strategyFamilyDecklist, fn: func() ([]gateway.Card, error) {
				t.Fatal("decklist-direct should not run after empty decklist-dedicated")
				return nil, nil
			}},
			fallbackAttempt{strategy: "scrap-dynamic", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
				t.Fatal("scrap-dynamic should not run after empty decklist-dedicated")
				return nil, nil
			}},
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 0 {
			t.Fatalf("expected zero cards, got %+v", cards)
		}
		if len(sequence) != 1 || sequence[0] != "decklist-dedicated" {
			t.Fatalf("expected only decklist-dedicated, got %v", sequence)
		}
	})

	t.Run("runs attempts in order and returns the final annotated error", func(t *testing.T) {
		sequence := make([]string, 0, 3)
		fail := func(label string) fallbackAttempt {
			return fallbackAttempt{
				strategy: label,
				family:   strategyFamilyFromName(label),
				fn: func() ([]gateway.Card, error) {
					sequence = append(sequence, label)
					return nil, fmt.Errorf("%s failed", label)
				},
			}
		}

		lastErr := errors.New("scrap-dynamic failed")
		_, err := runFallbackAttempts(
			fail("scrap-dedicated"),
			fail("scrap-direct"),
			fallbackAttempt{strategy: "scrap-dynamic", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
				sequence = append(sequence, "scrap-dynamic")
				return nil, lastErr
			}},
		)
		if !errors.Is(err, lastErr) {
			t.Fatalf("expected wrapped final error, got %v", err)
		}
		expectedError := "attempt 3 (scrap-dynamic): scrap-dynamic failed"
		if err == nil || err.Error() != expectedError {
			t.Fatalf("expected final error %q, got %v", expectedError, err)
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
}
