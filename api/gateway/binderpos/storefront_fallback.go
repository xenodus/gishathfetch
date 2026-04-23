package binderpos

import (
	"fmt"

	"mtg-price-checker-sg/gateway"
)

func searchWithFallback(
	searchAPIDedicatedFn func() ([]gateway.Card, error),
	searchAPISharedFn func() ([]gateway.Card, error),
	scrapDedicatedFn func() ([]gateway.Card, error),
	scrapSharedFn func() ([]gateway.Card, error),
) ([]gateway.Card, error) {
	// Some stores legitimately return no matches for the query.
	// If any attempt completes without an error, do not surface earlier failures.
	hasSuccessfulAttempt := false

	apiDedicatedCards, apiDedicatedErr := searchAPIDedicatedFn()
	apiDedicatedErr = annotateAttemptError(1, "api-dedicated", apiDedicatedErr)
	if apiDedicatedErr == nil {
		hasSuccessfulAttempt = true
	}
	if len(apiDedicatedCards) > 0 && apiDedicatedErr == nil {
		return apiDedicatedCards, nil
	}

	apiSharedCards, apiSharedErr := searchAPISharedFn()
	apiSharedErr = annotateAttemptError(2, "api-shared", apiSharedErr)
	if apiSharedErr == nil {
		hasSuccessfulAttempt = true
	}
	if len(apiSharedCards) > 0 && apiSharedErr == nil {
		return apiSharedCards, nil
	}

	scrapedCards, scrapErr := scrapDedicatedFn()
	scrapErr = annotateAttemptError(3, "scrap-dedicated", scrapErr)
	if scrapErr == nil {
		hasSuccessfulAttempt = true
	}
	if len(scrapedCards) > 0 && scrapErr == nil {
		return scrapedCards, nil
	}

	scrapedSharedCards, scrapSharedErr := scrapSharedFn()
	scrapSharedErr = annotateAttemptError(4, "scrap-shared", scrapSharedErr)
	if scrapSharedErr == nil {
		hasSuccessfulAttempt = true
	}
	if len(scrapedSharedCards) > 0 && scrapSharedErr == nil {
		return scrapedSharedCards, nil
	}

	if hasSuccessfulAttempt {
		return []gateway.Card{}, nil
	}

	if scrapSharedErr != nil {
		return scrapedSharedCards, scrapSharedErr
	}
	if scrapErr != nil {
		return scrapedCards, scrapErr
	}
	if apiSharedErr != nil {
		return apiSharedCards, apiSharedErr
	}
	if apiDedicatedErr != nil {
		return apiDedicatedCards, apiDedicatedErr
	}

	return []gateway.Card{}, nil
}

func annotateAttemptError(attempt int, strategy string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("attempt %d (%s): %w", attempt, strategy, err)
}
