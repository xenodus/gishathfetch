package binderpos

import (
	"context"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/config"
)

// ArcaneSanctumStoreName is the LGS name for Arcane Sanctum; referenced by store-specific gateway packages.
const ArcaneSanctumStoreName = "Arcane Sanctum"

// storefrontStrategy is a single ordered search attempt: a human-readable name
// used for error annotation plus the function that performs the lookup.
type storefrontStrategy struct {
	name string
	run  func(ctx context.Context) ([]gateway.Card, error)
}

func (i impl) Search(ctx context.Context, scrapVariant int, storeName, baseURL, shopifyDomain, searchURL, searchStr string, scrapOnly bool) ([]gateway.Card, error) {
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

	// ScrapOnly skips the shared decklist portal while keeping shopifyDomain
	// available for documentation and live integration tests. An empty domain
	// also forces scrape-only for backward compatibility.
	return runStorefrontStrategies(ctx, selectStorefrontStrategies(scrapOnly, shopifyDomain, scrap, decklist)...)
}

func selectStorefrontStrategies(scrapOnly bool, shopifyDomain string, scrap, decklist [3]storefrontStrategy) []storefrontStrategy {
	if scrapOnly || config.BinderposScrapOnly || strings.TrimSpace(shopifyDomain) == "" {
		return []storefrontStrategy{scrap[0], scrap[1], scrap[2]}
	}
	return orderDecklistAndScrap(decklist, scrap)
}

// orderDecklistAndScrap interleaves the decklist and scrap strategy families.
// The 50/50 lead decision (shouldStartWithDecklist) chooses which family is
// attempted first, so across the stores in one search roughly half lead with
// the shared BinderPOS decklist portal and half lead with their own storefront
// scrape. This halves the first-attempt burst on portal.binderpos.com, the host
// most prone to 429/503 throttling because every store would otherwise funnel
// into it at once. Within the ordering, the cheaper dedicated and direct
// attempts of both families run before either resorts to its dynamic-proxy
// attempt.
func orderDecklistAndScrap(decklist, scrap [3]storefrontStrategy) []storefrontStrategy {
	lead, follow := decklist, scrap
	if !shouldStartWithDecklist() {
		lead, follow = scrap, decklist
	}

	return []storefrontStrategy{
		lead[0], lead[1],
		follow[0], follow[1],
		lead[2], follow[2],
	}
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
