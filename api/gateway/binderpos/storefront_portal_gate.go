package binderpos

import (
	"context"

	"mtg-price-checker-sg/gateway"
)

// binderposPortalHost is the shared BinderPOS host that the decklist API funnels
// every store's primary lookup into. Because all BinderPOS-backed stores converge
// on this single host, it is throttled separately from per-store Shopify domains.
const binderposPortalHost = "portal.binderpos.com"

// binderposPortalMaxConcurrent caps how many decklist requests hit the shared
// portal host at once across all stores in a single search. The controller's
// binderposMaxConcurrent gate only limits concurrent stores, but every store's
// decklist call targets this one host, so an additional, smaller gate keyed on
// the host prevents bursts that trigger 429/503 responses.
const binderposPortalMaxConcurrent = 4

// binderposPortalGate bounds in-flight requests to the shared portal host.
var binderposPortalGate = make(chan struct{}, binderposPortalMaxConcurrent)

func init() {
	// Keep the shared portal host paced even when a store's first attempt opts
	// out of per-domain pacing for its own Shopify host.
	gateway.RegisterAlwaysPacedDomain(binderposPortalHost)
}

// acquireBinderposPortalSlot blocks until a slot on the shared portal host is
// free or ctx is done. The returned release must be called exactly once.
func acquireBinderposPortalSlot(ctx context.Context) (func(), error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case binderposPortalGate <- struct{}{}:
		return func() { <-binderposPortalGate }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
