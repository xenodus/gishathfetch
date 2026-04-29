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
	scrapDedicatedFn func() ([]gateway.Card, error),
	scrapDirectFn func() ([]gateway.Card, error),
) ([]gateway.Card, error) {
	apiDedicatedCards, apiDedicatedErr := searchAPIDedicatedFn()
	apiDedicatedErr = annotateAttemptError(1, "api-dedicated", apiDedicatedErr)
	if apiDedicatedErr == nil {
		return apiDedicatedCards, nil
	}

	scrapDedicatedCards, scrapDedicatedErr := scrapDedicatedFn()
	scrapDedicatedErr = annotateAttemptError(2, "scrap-dedicated", scrapDedicatedErr)
	if scrapDedicatedErr == nil {
		return scrapDedicatedCards, nil
	}

	scrapDirectCards, scrapDirectErr := scrapDirectFn()
	scrapDirectErr = annotateAttemptError(3, "scrap-direct", scrapDirectErr)
	if scrapDirectErr == nil {
		return scrapDirectCards, nil
	}

	// Reaching here means all three attempts errored.
	return scrapDirectCards, scrapDirectErr
}

func annotateAttemptError(attempt int, strategy string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("attempt %d (%s): %w", attempt, strategy, err)
}
