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
			fallbackAttempt{strategy: "decklist-dedicated", family: strategyFamilyDecklist, fn: func() ([]gateway.Card, error) {
				return []gateway.Card{{Name: "decklist-dedicated"}}, nil
			}},
			fallbackAttempt{strategy: "decklist-direct", family: strategyFamilyDecklist, fn: func() ([]gateway.Card, error) {
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
			fallbackAttempt{strategy: "decklist-dedicated", family: strategyFamilyDecklist, fn: func() ([]gateway.Card, error) {
				return nil, errors.New("decklist dedicated failed")
			}},
			fallbackAttempt{strategy: "scrap-dedicated", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
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

	t.Run("continues after an empty but error-free scrap attempt", func(t *testing.T) {
		cards, err := runFallbackAttempts(
			fallbackAttempt{strategy: "decklist-dedicated", family: strategyFamilyDecklist, fn: func() ([]gateway.Card, error) {
				return nil, errors.New("decklist dedicated failed")
			}},
			fallbackAttempt{strategy: "scrap-dedicated", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
				return []gateway.Card{}, nil
			}},
			fallbackAttempt{strategy: "scrap-direct", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
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

	t.Run("when decklist leads and first decklist is empty, skips remaining decklist and runs scrap", func(t *testing.T) {
		sequence := make([]string, 0, 4)
		record := func(label string, family strategyFamily) fallbackAttempt {
			return fallbackAttempt{
				strategy: label,
				family:   family,
				fn: func() ([]gateway.Card, error) {
					sequence = append(sequence, label)
					if label == "decklist-dedicated" {
						return []gateway.Card{}, nil
					}
					if label == "scrap-dedicated" {
						return []gateway.Card{{Name: "scrap-dedicated"}}, nil
					}
					t.Fatalf("unexpected attempt %q", label)
					return nil, nil
				},
			}
		}

		cards, err := runFallbackAttempts(
			record("decklist-dedicated", strategyFamilyDecklist),
			fallbackAttempt{strategy: "decklist-direct", family: strategyFamilyDecklist, fn: func() ([]gateway.Card, error) {
				t.Fatal("decklist-direct should be skipped after empty decklist-dedicated")
				return nil, nil
			}},
			record("scrap-dedicated", strategyFamilyScrap),
			fallbackAttempt{strategy: "decklist-dynamic", family: strategyFamilyDecklist, fn: func() ([]gateway.Card, error) {
				t.Fatal("decklist-dynamic should be skipped after empty decklist-dedicated")
				return nil, nil
			}},
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "scrap-dedicated" {
			t.Fatalf("expected scrap-dedicated card, got %+v", cards)
		}
		if len(sequence) != 2 || sequence[0] != "decklist-dedicated" || sequence[1] != "scrap-dedicated" {
			t.Fatalf("expected decklist-dedicated then scrap-dedicated, got %v", sequence)
		}
	})

	t.Run("when scrap leads and decklist is empty, still tries remaining decklist attempts", func(t *testing.T) {
		sequence := make([]string, 0, 4)
		cards, err := runFallbackAttempts(
			fallbackAttempt{strategy: "scrap-dedicated", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
				sequence = append(sequence, "scrap-dedicated")
				return []gateway.Card{}, nil
			}},
			fallbackAttempt{strategy: "scrap-direct", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
				sequence = append(sequence, "scrap-direct")
				return []gateway.Card{}, nil
			}},
			fallbackAttempt{strategy: "decklist-dedicated", family: strategyFamilyDecklist, fn: func() ([]gateway.Card, error) {
				sequence = append(sequence, "decklist-dedicated")
				return []gateway.Card{}, nil
			}},
			fallbackAttempt{strategy: "decklist-direct", family: strategyFamilyDecklist, fn: func() ([]gateway.Card, error) {
				sequence = append(sequence, "decklist-direct")
				return []gateway.Card{{Name: "decklist-direct"}}, nil
			}},
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(cards) != 1 || cards[0].Name != "decklist-direct" {
			t.Fatalf("expected decklist-direct card, got %+v", cards)
		}
		expected := []string{"scrap-dedicated", "scrap-direct", "decklist-dedicated", "decklist-direct"}
		if len(sequence) != len(expected) {
			t.Fatalf("expected %v, got %v", expected, sequence)
		}
		for i := range expected {
			if sequence[i] != expected[i] {
				t.Fatalf("attempt %d: expected %q, got %q", i+1, expected[i], sequence[i])
			}
		}
	})

	t.Run("returns empty only after the final attempt succeeds without cards", func(t *testing.T) {
		cards, err := runFallbackAttempts(
			fallbackAttempt{strategy: "decklist-dedicated", family: strategyFamilyDecklist, fn: func() ([]gateway.Card, error) {
				return []gateway.Card{}, nil
			}},
			fallbackAttempt{strategy: "scrap-direct", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
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
			return fallbackAttempt{
				strategy: label,
				family:   strategyFamilyFromName(label),
				fn: func() ([]gateway.Card, error) {
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
			fallbackAttempt{strategy: "scrap-dynamic", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
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
