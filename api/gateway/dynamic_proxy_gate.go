package gateway

import "context"

// dynamicProxyMaxConcurrent caps how many outbound requests may use DYNAMIC_PROXY
// at once across all stores and strategies. BinderPOS reserves dynamic proxy for
// the final scrap/decklist fallbacks, so many stores can reach those steps
// together after earlier strategies fail; without a shared gate the proxy
// endpoint is burst with concurrent traffic and often returns 429.
const dynamicProxyMaxConcurrent = 3

// dynamicProxyGate bounds in-flight requests routed through DYNAMIC_PROXY.
var dynamicProxyGate = make(chan struct{}, dynamicProxyMaxConcurrent)

// AcquireDynamicProxySlot blocks until a dynamic-proxy slot is free or ctx is
// done. The returned release must be called exactly once.
func AcquireDynamicProxySlot(ctx context.Context) (func(), error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case dynamicProxyGate <- struct{}{}:
		return func() { <-dynamicProxyGate }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
