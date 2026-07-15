package binderpos

import (
	"context"

	"mtg-price-checker-sg/gateway"
)

const (
	// scrapDynamicMaxAttempts bounds how many times a single scrap-dynamic call
	// is sent when the upstream responds with 429. Each retry uses a fresh
	// collector so the rotating proxy egresses from a new IP.
	scrapDynamicMaxAttempts = 3
)

func (i impl) scrapDynamic(ctx context.Context, scrapVariant int, storeName, baseUrl, searchUrl, searchStr string) ([]gateway.Card, error) {
	releaseDynamicProxy, err := gateway.AcquireDynamicProxySlot(ctx)
	if err != nil {
		return nil, err
	}
	defer releaseDynamicProxy()

	return scrapDynamicWithRetry(ctx, func() ([]gateway.Card, error) {
		return i.scrapWithCollectorFactory(
			ctx,
			scrapVariant,
			storeName,
			baseUrl,
			searchUrl,
			searchStr,
			newDynamicNoRetryCollector,
		)
	})
}

func scrapDynamicWithRetry(ctx context.Context, run func() ([]gateway.Card, error)) ([]gateway.Card, error) {
	var (
		cards   []gateway.Card
		lastErr error
	)

	for attempt := range scrapDynamicMaxAttempts {
		cards, lastErr = run()
		if lastErr == nil {
			return cards, nil
		}
		if !gateway.IsHTTPTooManyRequests(lastErr) || isLastScrapDynamicAttempt(attempt) {
			return cards, lastErr
		}
		if !waitBeforeDecklistRetry(ctx, decklistBackoffDelay(attempt)) {
			return cards, lastErr
		}
	}

	return cards, lastErr
}

func isLastScrapDynamicAttempt(attempt int) bool {
	return attempt >= scrapDynamicMaxAttempts-1
}
