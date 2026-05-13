package binderpos

import (
	"fmt"

	"mtg-price-checker-sg/gateway"
)

type fallbackAttempt struct {
	strategy string
	fn       func() ([]gateway.Card, error)
}

// searchWithScrapDedicatedThenDirectThenDynamic tries a scrape using dedicated-proxy routing first,
// then a direct (no proxy) scrape, and only uses the dynamic proxy as the last fallback.
func searchWithScrapDedicatedThenDirectThenDynamic(
	scrapDedicatedFn func() ([]gateway.Card, error),
	scrapDirectFn func() ([]gateway.Card, error),
	scrapDynamicFn func() ([]gateway.Card, error),
) ([]gateway.Card, error) {
	return runFallbackAttempts(
		fallbackAttempt{strategy: "scrap-dedicated", fn: scrapDedicatedFn},
		fallbackAttempt{strategy: "scrap-direct", fn: scrapDirectFn},
		fallbackAttempt{strategy: "scrap-dynamic", fn: scrapDynamicFn},
	)
}

func searchWithFallback(
	searchAPIDedicatedFn func() ([]gateway.Card, error),
	searchAPIDirectFn func() ([]gateway.Card, error),
	scrapDedicatedFn func() ([]gateway.Card, error),
	scrapDirectFn func() ([]gateway.Card, error),
	searchAPIDynamicFn func() ([]gateway.Card, error),
	scrapDynamicFn func() ([]gateway.Card, error),
) ([]gateway.Card, error) {
	return runFallbackAttempts(
		fallbackAttempt{strategy: "api-dedicated", fn: searchAPIDedicatedFn},
		fallbackAttempt{strategy: "api-direct", fn: searchAPIDirectFn},
		fallbackAttempt{strategy: "scrap-dedicated", fn: scrapDedicatedFn},
		fallbackAttempt{strategy: "scrap-direct", fn: scrapDirectFn},
		fallbackAttempt{strategy: "api-dynamic", fn: searchAPIDynamicFn},
		fallbackAttempt{strategy: "scrap-dynamic", fn: scrapDynamicFn},
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
