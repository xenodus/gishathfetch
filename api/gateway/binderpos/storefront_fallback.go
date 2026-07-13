package binderpos

import (
	"fmt"

	"mtg-price-checker-sg/gateway"
)

type fallbackAttempt struct {
	strategy string
	fn       func() ([]gateway.Card, error)
}

// runFallbackAttempts runs the supplied attempts in order, returning the first
// result that returns cards. An attempt that finishes without error but finds no
// cards is treated as inconclusive and the next strategy is tried; only the
// final attempt may return an empty result. Each attempt's error is annotated
// with its position and strategy name so the final error reflects the last
// attempt tried.
func runFallbackAttempts(attempts ...fallbackAttempt) ([]gateway.Card, error) {
	var (
		cards []gateway.Card
		err   error
	)

	for idx, attempt := range attempts {
		cards, err = attempt.fn()
		err = annotateAttemptError(idx+1, attempt.strategy, err)
		if err == nil && (len(cards) > 0 || idx == len(attempts)-1) {
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
