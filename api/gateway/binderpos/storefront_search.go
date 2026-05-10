package binderpos

import (
	"context"
	"strings"

	"mtg-price-checker-sg/gateway"
)

// ArcaneSanctumStoreName is the LGS name for Arcane Sanctum; referenced by store-specific gateway packages.
const ArcaneSanctumStoreName = "Arcane Sanctum"

func (i impl) Search(ctx context.Context, scrapVariant int, storeName, baseURL, shopifyDomain, searchURL, searchStr string) ([]gateway.Card, error) {
	if strings.TrimSpace(shopifyDomain) == "" {
		return searchWithScrapDedicatedThenDynamicThenDirect(
			func() ([]gateway.Card, error) {
				return runWithAttemptTimeout(ctx, func(attemptCtx context.Context) ([]gateway.Card, error) {
					return i.Scrap(attemptCtx, scrapVariant, storeName, baseURL, searchURL, searchStr)
				})
			},
			func() ([]gateway.Card, error) {
				return runWithAttemptTimeout(ctx, func(attemptCtx context.Context) ([]gateway.Card, error) {
					return i.scrapDynamic(attemptCtx, scrapVariant, storeName, baseURL, searchURL, searchStr)
				})
			},
			func() ([]gateway.Card, error) {
				return runWithAttemptTimeout(ctx, func(attemptCtx context.Context) ([]gateway.Card, error) {
					return i.scrapDirect(attemptCtx, scrapVariant, storeName, baseURL, searchURL, searchStr)
				})
			},
		)
	}

	return searchWithFallback(
		func() ([]gateway.Card, error) {
			return runWithAttemptTimeout(ctx, func(attemptCtx context.Context) ([]gateway.Card, error) {
				return searchByStorefrontAPI(attemptCtx, scrapVariant, storeName, baseURL, shopifyDomain, searchStr)
			})
		},
		func() ([]gateway.Card, error) {
			return runWithAttemptTimeout(ctx, func(attemptCtx context.Context) ([]gateway.Card, error) {
				return searchByStorefrontAPIDynamic(attemptCtx, scrapVariant, storeName, baseURL, shopifyDomain, searchStr)
			})
		},
		func() ([]gateway.Card, error) {
			return runWithAttemptTimeout(ctx, func(attemptCtx context.Context) ([]gateway.Card, error) {
				return i.Scrap(attemptCtx, scrapVariant, storeName, baseURL, searchURL, searchStr)
			})
		},
		func() ([]gateway.Card, error) {
			return runWithAttemptTimeout(ctx, func(attemptCtx context.Context) ([]gateway.Card, error) {
				return i.scrapDynamic(attemptCtx, scrapVariant, storeName, baseURL, searchURL, searchStr)
			})
		},
		func() ([]gateway.Card, error) {
			return runWithAttemptTimeout(ctx, func(attemptCtx context.Context) ([]gateway.Card, error) {
				return i.scrapDirect(attemptCtx, scrapVariant, storeName, baseURL, searchURL, searchStr)
			})
		},
	)
}

func runWithAttemptTimeout(ctx context.Context, fn func(context.Context) ([]gateway.Card, error)) ([]gateway.Card, error) {
	attemptCtx, cancel := context.WithTimeout(ctx, binderposAttemptTimeout)
	defer cancel()
	return fn(attemptCtx)
}
