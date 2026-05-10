package binderpos

import (
	"fmt"

	"mtg-price-checker-sg/gateway"
)

type fallbackAttempt struct {
	strategy string
	fn       func() ([]gateway.Card, error)
}

// searchWithScrapDedicatedThenDynamicThenDirect tries a scrape using dedicated-proxy routing first,
// then dynamic proxy routing, then a direct (no proxy) scrape on failure.
func searchWithScrapDedicatedThenDynamicThenDirect(
	scrapDedicatedFn func() ([]gateway.Card, error),
	scrapDynamicFn func() ([]gateway.Card, error),
	scrapDirectFn func() ([]gateway.Card, error),
) ([]gateway.Card, error) {
	return runFallbackAttempts(
		fallbackAttempt{strategy: "scrap-dedicated", fn: scrapDedicatedFn},
		fallbackAttempt{strategy: "scrap-dynamic", fn: scrapDynamicFn},
		fallbackAttempt{strategy: "scrap-direct", fn: scrapDirectFn},
	)
}

func searchWithFallback(
	searchAPIDedicatedFn func() ([]gateway.Card, error),
	searchAPIDynamicFn func() ([]gateway.Card, error),
	scrapDedicatedFn func() ([]gateway.Card, error),
	scrapDynamicFn func() ([]gateway.Card, error),
	scrapDirectFn func() ([]gateway.Card, error),
) ([]gateway.Card, error) {
	return runFallbackAttempts(
		fallbackAttempt{strategy: "api-dedicated", fn: searchAPIDedicatedFn},
		fallbackAttempt{strategy: "api-dynamic", fn: searchAPIDynamicFn},
		fallbackAttempt{strategy: "scrap-dedicated", fn: scrapDedicatedFn},
		fallbackAttempt{strategy: "scrap-dynamic", fn: scrapDynamicFn},
		fallbackAttempt{strategy: "scrap-direct", fn: scrapDirectFn},
	)
}

func runFallbackAttempts(attempts ...fallbackAttempt) ([]gateway.Card, error) {
	var (
		cards []gateway.Card
		err   error
	)

	for idx, attempt := range attempts {
		cards, err = attempt.fn()
		err = annotateAttemptError(idx+1, attempt.strategy, err)
		if err == nil {
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
