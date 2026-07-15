package binderpos

import (
	"fmt"
	"strings"

	"mtg-price-checker-sg/gateway"
)

type strategyFamily int

const (
	strategyFamilyUnknown strategyFamily = iota
	strategyFamilyDecklist
	strategyFamilyScrap
)

type fallbackAttempt struct {
	strategy string
	family   strategyFamily
	fn       func() ([]gateway.Card, error)
}

func strategyFamilyFromName(name string) strategyFamily {
	switch {
	case strings.HasPrefix(name, "decklist-"):
		return strategyFamilyDecklist
	case strings.HasPrefix(name, "scrap-"):
		return strategyFamilyScrap
	default:
		return strategyFamilyUnknown
	}
}

// runFallbackAttempts runs the supplied attempts in order, returning the first
// result that returns cards.
//
// When any scrap attempt returns no cards without error, that empty result is
// final and no further strategies run. When any decklist attempt returns no
// cards without error, remaining decklist attempts are skipped. HTTP 5xx errors
// on scrap attempts are final so a failing storefront is not followed by the
// shared portal. Each attempt's error is annotated with its position and
// strategy name so the final error reflects the last attempt tried.
func runFallbackAttempts(attempts ...fallbackAttempt) ([]gateway.Card, error) {
	var (
		cards           []gateway.Card
		err             error
		executedAttempt int
	)

	abandonDecklist := false

	for _, attempt := range attempts {
		if abandonDecklist && attempt.family == strategyFamilyDecklist {
			continue
		}

		executedAttempt++
		cards, err = attempt.fn()
		err = annotateAttemptError(executedAttempt, attempt.strategy, err)

		if err == nil && len(cards) > 0 {
			return cards, nil
		}

		if attempt.family == strategyFamilyScrap && err == nil && len(cards) == 0 {
			return cards, nil
		}

		if attempt.family == strategyFamilyDecklist && err == nil && len(cards) == 0 {
			abandonDecklist = true
			continue
		}

		if attempt.family == strategyFamilyScrap && gateway.IsHTTPServerError(err) {
			return cards, err
		}
	}

	return cards, err
}

func annotateAttemptError(attempt int, strategy string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("attempt %d (%s): %w", attempt, strategy, err)
}
