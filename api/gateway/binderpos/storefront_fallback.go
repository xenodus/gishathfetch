package binderpos

import (
	"fmt"

	"mtg-price-checker-sg/gateway"
)

// searchWithScrapDedicatedThenDirect tries a scrape using dedicated-proxy routing first (Scrap), then
// a direct (no proxy) scrape on failure.
func searchWithScrapDedicatedThenDirect(
	scrapDedicatedFn func() ([]gateway.Card, error),
	scrapDirectFn func() ([]gateway.Card, error),
) ([]gateway.Card, error) {
	cards, dedicatedErr := scrapDedicatedFn()
	dedicatedErr = annotateAttemptError(1, "scrap-dedicated", dedicatedErr)
	if dedicatedErr == nil {
		return cards, nil
	}

	cards, directErr := scrapDirectFn()
	directErr = annotateAttemptError(2, "scrap-direct", directErr)
	if directErr == nil {
		return cards, nil
	}
	return cards, directErr
}

func searchWithFallback(
	searchAPIDedicatedFn func() ([]gateway.Card, error),
	searchAPISharedFn func() ([]gateway.Card, error),
	scrapSharedFn func() ([]gateway.Card, error),
) ([]gateway.Card, error) {
	apiDedicatedCards, apiDedicatedErr := searchAPIDedicatedFn()
	apiDedicatedErr = annotateAttemptError(1, "api-dedicated", apiDedicatedErr)
	if apiDedicatedErr == nil {
		return apiDedicatedCards, nil
	}

	apiSharedCards, apiSharedErr := searchAPISharedFn()
	apiSharedErr = annotateAttemptError(2, "api-shared", apiSharedErr)
	if apiSharedErr == nil {
		return apiSharedCards, nil
	}

	scrapedSharedCards, scrapSharedErr := scrapSharedFn()
	scrapSharedErr = annotateAttemptError(3, "scrap-shared", scrapSharedErr)
	if scrapSharedErr == nil {
		return scrapedSharedCards, nil
	}

	// Reaching here means all three attempts errored.
	return scrapedSharedCards, scrapSharedErr
}

func annotateAttemptError(attempt int, strategy string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("attempt %d (%s): %w", attempt, strategy, err)
}
