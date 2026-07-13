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
// result that returns cards.
//
// When any attempt returns no cards without error, that empty result is final
// and no further strategies run. Each attempt's error is annotated with its
// position and strategy name so the final error reflects the last attempt tried.
func runFallbackAttempts(attempts ...fallbackAttempt) ([]gateway.Card, error) {
	var (
		cards           []gateway.Card
		err             error
		executedAttempt int
	)

	for _, attempt := range attempts {
		executedAttempt++
		cards, err = attempt.fn()
		err = annotateAttemptError(executedAttempt, attempt.strategy, err)

		if err == nil && len(cards) > 0 {
			return cards, nil
		}

		if err == nil && len(cards) == 0 {
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
