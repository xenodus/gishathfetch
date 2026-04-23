package binderpos

import (
	"fmt"

	"mtg-price-checker-sg/gateway"
)

func searchWithFallback(
	searchAPIDedicatedFn func() ([]gateway.Card, error),
	searchAPISharedFn func() ([]gateway.Card, error),
	scrapSharedPrimaryFn func() ([]gateway.Card, error),
	scrapSharedSecondaryFn func() ([]gateway.Card, error),
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

	scrapedCards, scrapErr := scrapSharedPrimaryFn()
	scrapErr = annotateAttemptError(3, "scrap-shared-primary", scrapErr)
	if scrapErr == nil {
		return scrapedCards, nil
	}

	scrapedSharedCards, scrapSharedErr := scrapSharedSecondaryFn()
	scrapSharedErr = annotateAttemptError(4, "scrap-shared-secondary", scrapSharedErr)
	if scrapSharedErr == nil {
		return scrapedSharedCards, nil
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
