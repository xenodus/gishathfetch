package binderpos

import (
	"context"
	"strings"
	"time"

	"mtg-price-checker-sg/gateway"
)

// storefrontStrategy is a single ordered search attempt: a human-readable name
// used for error annotation plus the function that performs the lookup.
type storefrontStrategy struct {
	name string
	run  func(ctx context.Context) ([]gateway.Card, error)
}

func (i impl) Search(ctx context.Context, scrapVariant int, storeName, baseURL, shopifyDomain, searchURL, searchStr, storefrontAccessToken string) ([]gateway.Card, error) {
	scrap := [2]storefrontStrategy{
		{
			name: "scrap-direct",
			run: func(attemptCtx context.Context) ([]gateway.Card, error) {
				return i.scrapDirect(attemptCtx, scrapVariant, storeName, baseURL, searchURL, searchStr)
			},
		},
		{
			name: "scrap-dedicated",
			run: func(attemptCtx context.Context) ([]gateway.Card, error) {
				return i.Scrap(attemptCtx, scrapVariant, storeName, baseURL, searchURL, searchStr)
			},
		},
	}

	var strategies []storefrontStrategy
	if token := strings.TrimSpace(storefrontAccessToken); token != "" {
		strategies = append(strategies,
			storefrontStrategy{
				name: "graphql-direct",
				run: func(attemptCtx context.Context) ([]gateway.Card, error) {
					return searchByStorefrontGraphQLDirect(attemptCtx, scrapVariant, storeName, baseURL, token, searchStr)
				},
			},
			storefrontStrategy{
				name: "graphql-dedicated",
				run: func(attemptCtx context.Context) ([]gateway.Card, error) {
					return searchByStorefrontGraphQLDedicated(attemptCtx, scrapVariant, storeName, baseURL, token, searchStr)
				},
			},
		)
	}

	strategies = append(strategies, scrap[0], scrap[1])

	return runStorefrontStrategies(ctx, omitDedicatedStorefrontStrategies(strategies)...)
}

func omitDedicatedStorefrontStrategies(strategies []storefrontStrategy) []storefrontStrategy {
	if gateway.DedicatedProxiesEnabled() {
		return strategies
	}
	filtered := make([]storefrontStrategy, 0, len(strategies))
	for _, strategy := range strategies {
		if strings.HasSuffix(strategy.name, "-dedicated") {
			continue
		}
		filtered = append(filtered, strategy)
	}
	return filtered
}

// runStorefrontStrategies runs the ordered strategies through the shared
// fallback runner. The first attempt starts immediately; later attempts honor
// per-domain request pacing. Each attempt is bounded by a transport-specific timeout.
func runStorefrontStrategies(ctx context.Context, strategies ...storefrontStrategy) ([]gateway.Card, error) {
	attempts := make([]fallbackAttempt, len(strategies))
	for idx := range strategies {
		strategy := strategies[idx]
		applyRequestPacing := idx != 0
		attempts[idx] = fallbackAttempt{
			strategy: strategy.name,
			family:   strategyFamilyFromName(strategy.name),
			fn: func() ([]gateway.Card, error) {
				return runWithAttemptTimeout(ctx, applyRequestPacing, attemptTimeoutForStrategy(strategy.name), strategy.run)
			},
		}
	}

	return runFallbackAttempts(attempts...)
}

func attemptTimeoutForStrategy(name string) time.Duration {
	if strings.HasSuffix(name, "-direct") {
		return binderposDirectAttemptTimeout
	}
	return binderposDedicatedAttemptTimeout
}

func runWithAttemptTimeout(ctx context.Context, applyRequestPacing bool, attemptTimeout time.Duration, fn func(context.Context) ([]gateway.Card, error)) ([]gateway.Card, error) {
	if !applyRequestPacing {
		// Let the first BinderPOS attempt start immediately; fallbacks still share pacing.
		ctx = gateway.WithDomainRequestPacingDisabled(ctx)
	}
	attemptCtx, cancel := context.WithTimeout(ctx, attemptTimeout)
	defer cancel()
	return fn(attemptCtx)
}

// storefrontStrategyNames returns the ordered strategy names for the given
// storefront token and Shopify domain. Used by tests.
func storefrontStrategyNames(storefrontAccessToken string) []string {
	return omitDedicatedStrategyNames(storefrontStrategyNamesUnfiltered(storefrontAccessToken))
}

func storefrontStrategyNamesUnfiltered(storefrontAccessToken string) []string {
	var names []string
	if strings.TrimSpace(storefrontAccessToken) != "" {
		names = append(names, "graphql-direct", "graphql-dedicated")
	}
	names = append(names, "scrap-direct", "scrap-dedicated")
	return names
}

func omitDedicatedStrategyNames(names []string) []string {
	if gateway.DedicatedProxiesEnabled() {
		return names
	}
	filtered := make([]string, 0, len(names))
	for _, name := range names {
		if strings.HasSuffix(name, "-dedicated") {
			continue
		}
		filtered = append(filtered, name)
	}
	return filtered
}
