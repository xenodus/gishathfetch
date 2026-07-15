package binderpos

import (
	"context"
	"strings"

	"mtg-price-checker-sg/gateway"
)

// storefrontStrategy is a single ordered search attempt: a human-readable name
// used for error annotation plus the function that performs the lookup.
type storefrontStrategy struct {
	name string
	run  func(ctx context.Context) ([]gateway.Card, error)
}

func (i impl) Search(ctx context.Context, scrapVariant int, storeName, baseURL, shopifyDomain, searchURL, searchStr, storefrontAccessToken string) ([]gateway.Card, error) {
	scrap := [3]storefrontStrategy{
		{
			name: "scrap-dedicated",
			run: func(attemptCtx context.Context) ([]gateway.Card, error) {
				return i.Scrap(attemptCtx, scrapVariant, storeName, baseURL, searchURL, searchStr)
			},
		},
		{
			name: "scrap-direct",
			run: func(attemptCtx context.Context) ([]gateway.Card, error) {
				return i.scrapDirect(attemptCtx, scrapVariant, storeName, baseURL, searchURL, searchStr)
			},
		},
		{
			name: "scrap-dynamic",
			run: func(attemptCtx context.Context) ([]gateway.Card, error) {
				return i.scrapDynamic(attemptCtx, scrapVariant, storeName, baseURL, searchURL, searchStr)
			},
		},
	}

	var strategies []storefrontStrategy
	if token := strings.TrimSpace(storefrontAccessToken); token != "" {
		strategies = append(strategies,
			storefrontStrategy{
				name: "graphql-dedicated",
				run: func(attemptCtx context.Context) ([]gateway.Card, error) {
					return searchByStorefrontGraphQLDedicated(attemptCtx, scrapVariant, storeName, baseURL, token, searchStr)
				},
			},
			storefrontStrategy{
				name: "graphql-direct",
				run: func(attemptCtx context.Context) ([]gateway.Card, error) {
					return searchByStorefrontGraphQLDirect(attemptCtx, scrapVariant, storeName, baseURL, token, searchStr)
				},
			},
		)
	}

	strategies = append(strategies, scrap[0], scrap[1])
	var decklistDynamic storefrontStrategy
	if strings.TrimSpace(shopifyDomain) != "" {
		decklist := [3]storefrontStrategy{
			{
				name: "decklist-dedicated",
				run: func(attemptCtx context.Context) ([]gateway.Card, error) {
					return searchByStorefrontAPI(attemptCtx, scrapVariant, storeName, baseURL, shopifyDomain, searchStr)
				},
			},
			{
				name: "decklist-direct",
				run: func(attemptCtx context.Context) ([]gateway.Card, error) {
					return searchByStorefrontAPIDirect(attemptCtx, scrapVariant, storeName, baseURL, shopifyDomain, searchStr)
				},
			},
			{
				name: "decklist-dynamic",
				run: func(attemptCtx context.Context) ([]gateway.Card, error) {
					return searchByStorefrontAPIDynamic(attemptCtx, scrapVariant, storeName, baseURL, shopifyDomain, searchStr)
				},
			},
		}
		strategies = append(strategies, decklist[0], decklist[1])
		decklistDynamic = decklist[2]
	}
	strategies = append(strategies, scrap[2])
	if strings.TrimSpace(shopifyDomain) != "" {
		strategies = append(strategies, decklistDynamic)
	}

	return runStorefrontStrategies(ctx, strategies...)
}

// runStorefrontStrategies runs the ordered strategies through the shared
// fallback runner. The first attempt starts immediately; later attempts honor
// per-domain request pacing. Each attempt is bounded by binderposAttemptTimeout.
func runStorefrontStrategies(ctx context.Context, strategies ...storefrontStrategy) ([]gateway.Card, error) {
	attempts := make([]fallbackAttempt, len(strategies))
	for idx := range strategies {
		strategy := strategies[idx]
		applyRequestPacing := idx != 0
		attempts[idx] = fallbackAttempt{
			strategy: strategy.name,
			family:   strategyFamilyFromName(strategy.name),
			fn: func() ([]gateway.Card, error) {
				return runWithAttemptTimeout(ctx, applyRequestPacing, strategy.run)
			},
		}
	}

	return runFallbackAttempts(attempts...)
}

func runWithAttemptTimeout(ctx context.Context, applyRequestPacing bool, fn func(context.Context) ([]gateway.Card, error)) ([]gateway.Card, error) {
	if !applyRequestPacing {
		// Let the first BinderPOS attempt start immediately; fallbacks still share pacing.
		ctx = gateway.WithDomainRequestPacingDisabled(ctx)
	}
	attemptCtx, cancel := context.WithTimeout(ctx, binderposAttemptTimeout)
	defer cancel()
	return fn(attemptCtx)
}

// storefrontStrategyNames returns the ordered strategy names for the given
// storefront token and Shopify domain. Used by tests.
func storefrontStrategyNames(storefrontAccessToken, shopifyDomain string) []string {
	var names []string
	if strings.TrimSpace(storefrontAccessToken) != "" {
		names = append(names, "graphql-dedicated", "graphql-direct")
	}
	names = append(names, "scrap-dedicated", "scrap-direct")
	if strings.TrimSpace(shopifyDomain) != "" {
		names = append(names, "decklist-dedicated", "decklist-direct")
	}
	names = append(names, "scrap-dynamic")
	if strings.TrimSpace(shopifyDomain) != "" {
		names = append(names, "decklist-dynamic")
	}
	return names
}
