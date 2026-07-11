package shopifysuggest

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
)

// shopifySuggestDedicatedProxySeq advances once per dedicated-proxy suggest
// attempt so traffic round-robins evenly across the configured dedicated
// proxies instead of always hammering the first one.
var shopifySuggestDedicatedProxySeq atomic.Uint32

// searchAttempt pairs a human-readable strategy name (used for error
// annotation) with the HTTP client that performs that attempt.
type searchAttempt struct {
	strategy string
	client   *http.Client
}

// searchWithProxyFallback runs the suggest request through an ordered set of
// transports, returning the first successful result. Each transport retries
// transient rate-limit/5xx responses on the same connection (honoring
// Retry-After) before the chain advances. Dedicated proxies are tried first so
// cloud/datacenter egress is less likely to trip Shopify throttling; direct and
// dynamic transports follow when dedicated is unavailable or exhausted.
func searchWithProxyFallback(ctx context.Context, opts Options, apiURL string) ([]gateway.Card, error) {
	return runSearchAttempts(ctx, buildSearchAttempts(), opts, apiURL)
}

// runSearchAttempts executes the ordered attempts, returning the first result
// that succeeds (a nil error, even with zero cards). Each attempt's error is
// annotated with its position and strategy name so the final error reflects the
// last transport tried.
func runSearchAttempts(ctx context.Context, attempts []searchAttempt, opts Options, apiURL string) ([]gateway.Card, error) {
	var (
		cards []gateway.Card
		err   error
	)
	for idx, attempt := range attempts {
		cards, err = fetchAndMapProducts(ctx, attempt.client, apiURL, opts)
		err = annotateSuggestAttemptError(idx+1, attempt.strategy, err)
		if err == nil {
			return cards, nil
		}
	}
	return cards, err
}

// buildSearchAttempts builds the ordered fallback chain: dedicated proxy (when
// configured), then direct, then dynamic proxy (when configured).
func buildSearchAttempts() []searchAttempt {
	var attempts []searchAttempt

	if client, ok := dedicatedProxyClient(); ok {
		attempts = append(attempts, searchAttempt{strategy: "dedicated", client: client})
	}

	attempts = append(attempts, searchAttempt{
		strategy: "direct",
		client:   &http.Client{Timeout: config.SearchAttemptTimeout},
	})

	if client, ok := dynamicProxyClient(); ok {
		attempts = append(attempts, searchAttempt{strategy: "dynamic", client: client})
	}

	return attempts
}

// productDetailClient selects the HTTP client for a product JSON fetch. Tests may
// replace this to inject a local client.
var productDetailClient = defaultProductDetailClient

// defaultProductDetailClient returns a fresh client for each product detail
// request. Dedicated proxies are round-robined so variant resolution spreads
// load across IPs instead of reusing the transport that won the suggest
// fallback. When no dedicated proxies are configured, direct is used, then
// dynamic when available.
func defaultProductDetailClient() *http.Client {
	if client, ok := dedicatedProxyClient(); ok {
		return client
	}
	if client, ok := dynamicProxyClient(); ok {
		return client
	}
	return &http.Client{Timeout: config.SearchAttemptTimeout}
}

// dedicatedProxyClient returns an HTTP client bound to the next dedicated proxy
// in round-robin order, or ok=false when none are configured.
func dedicatedProxyClient() (*http.Client, bool) {
	proxyURLs := util.GetDedicatedProxyURLs()
	if len(proxyURLs) == 0 {
		return nil, false
	}

	client, err := newProxyClient(nextDedicatedProxyURL(proxyURLs))
	if err != nil {
		return nil, false
	}
	return client, true
}

// dynamicProxyClient returns an HTTP client bound to the shared dynamic proxy,
// or ok=false when it is not configured.
func dynamicProxyClient() (*http.Client, bool) {
	proxyURL := gateway.DynamicProxyURL()
	if proxyURL == "" {
		return nil, false
	}

	client, err := newProxyClient(proxyURL)
	if err != nil {
		return nil, false
	}
	return client, true
}

// nextDedicatedProxyURL returns the next dedicated proxy URL in round-robin order.
func nextDedicatedProxyURL(proxyURLs []string) string {
	v := shopifySuggestDedicatedProxySeq.Add(1) - 1
	return proxyURLs[int(v%uint32(len(proxyURLs)))]
}

func newProxyClient(proxyURL string) (*http.Client, error) {
	parsedProxyURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout: config.SearchAttemptTimeout,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsedProxyURL),
		},
	}, nil
}

func fetchAndMapProducts(ctx context.Context, client *http.Client, apiURL string, opts Options) ([]gateway.Card, error) {
	products, err := fetchProducts(ctx, client, apiURL, suggestRequestOptsFromConfig(opts.Config))
	if err != nil {
		return nil, err
	}
	return mapProducts(ctx, opts, products), nil
}

func annotateSuggestAttemptError(attempt int, strategy string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("attempt %d (%s): %w", attempt, strategy, err)
}
