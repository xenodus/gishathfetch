package gateway

import (
	"context"
	"strings"
)

type requestDedicatedProxyKey struct{}

// WithRequestDedicatedProxy pins one dedicated proxy URL on ctx for a single store
// search. The caller must hold the lease separately and release it when the store
// search finishes.
func WithRequestDedicatedProxy(ctx context.Context, proxyURL string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	proxyURL = strings.TrimSpace(proxyURL)
	if proxyURL == "" {
		return ctx
	}
	return context.WithValue(ctx, requestDedicatedProxyKey{}, proxyURL)
}

// RequestDedicatedProxyURL returns the dedicated proxy URL pinned on ctx, if any.
func RequestDedicatedProxyURL(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	proxyURL, ok := ctx.Value(requestDedicatedProxyKey{}).(string)
	if !ok || strings.TrimSpace(proxyURL) == "" {
		return "", false
	}
	return proxyURL, true
}
