package gateway

import "context"

// DedicatedProxySearchMaxConcurrent caps how many store searches may hold a
// dedicated-proxy lease at once. The controller runs more workers than this,
// but proxy-backed searches queue here so datacenter egress does not burst
// every configured DEDICATED_PROXY_* slot at once.
const DedicatedProxySearchMaxConcurrent = 3

// dedicatedProxySearchGate bounds in-flight store searches that use dedicated proxies.
var dedicatedProxySearchGate = make(chan struct{}, DedicatedProxySearchMaxConcurrent)

// AcquireDedicatedProxySearchSlot blocks until a dedicated-proxy search slot is
// free or ctx is done. The returned release must be called exactly once.
func AcquireDedicatedProxySearchSlot(ctx context.Context) (func(), error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case dedicatedProxySearchGate <- struct{}{}:
		return func() { <-dedicatedProxySearchGate }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
