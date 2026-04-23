package binderpos

import (
	"context"

	"mtg-price-checker-sg/gateway"
)

func (i impl) Search(ctx context.Context, scrapVariant int, storeName, baseURL, searchURL, searchStr string) ([]gateway.Card, error) {
	return searchWithFallback(
		func() ([]gateway.Card, error) {
			return runWithAttemptTimeout(ctx, func(attemptCtx context.Context) ([]gateway.Card, error) {
				return searchByStorefrontAPI(attemptCtx, scrapVariant, storeName, baseURL, searchStr)
			})
		},
		func() ([]gateway.Card, error) {
			return runWithAttemptTimeout(ctx, func(attemptCtx context.Context) ([]gateway.Card, error) {
				return searchByStorefrontAPISharedProxy(attemptCtx, scrapVariant, storeName, baseURL, searchStr)
			})
		},
		func() ([]gateway.Card, error) {
			return runWithAttemptTimeout(ctx, func(attemptCtx context.Context) ([]gateway.Card, error) {
				return i.scrapDedicatedProxy(attemptCtx, scrapVariant, storeName, baseURL, searchURL, searchStr)
			})
		},
		func() ([]gateway.Card, error) {
			return runWithAttemptTimeout(ctx, func(attemptCtx context.Context) ([]gateway.Card, error) {
				return i.scrapSharedProxy(attemptCtx, scrapVariant, storeName, baseURL, searchURL, searchStr)
			})
		},
	)
}

func runWithAttemptTimeout(ctx context.Context, fn func(context.Context) ([]gateway.Card, error)) ([]gateway.Card, error) {
	attemptCtx, cancel := context.WithTimeout(ctx, binderposAttemptTimeout)
	defer cancel()
	return fn(attemptCtx)
}
