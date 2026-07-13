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
// result that returns cards. An attempt that finishes without error but finds no
// cards is treated as inconclusive and the next strategy is tried; only the
// final attempt may return an empty result.
//
// When decklist leads and the first decklist attempt returns no cards without
// error, remaining decklist attempts are skipped and scrap strategies run next.
// Each attempt's error is annotated with its position and strategy name so the
// final error reflects the last attempt tried.
func runFallbackAttempts(attempts ...fallbackAttempt) ([]gateway.Card, error) {
	var (
		cards           []gateway.Card
		err             error
		executedAttempt int
	)

	decklistLeads := len(attempts) > 0 && attempts[0].family == strategyFamilyDecklist
	abandonDecklist := false

	for idx, attempt := range attempts {
		if abandonDecklist && attempt.family == strategyFamilyDecklist {
			continue
		}

		executedAttempt++
		cards, err = attempt.fn()
		err = annotateAttemptError(executedAttempt, attempt.strategy, err)

		if err == nil && len(cards) > 0 {
			return cards, nil
		}

		if decklistLeads && !abandonDecklist &&
			attempt.family == strategyFamilyDecklist && err == nil && len(cards) == 0 {
			abandonDecklist = true
			continue
		}

		if err == nil && len(cards) == 0 && idx == len(attempts)-1 {
			return cards, nil
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
