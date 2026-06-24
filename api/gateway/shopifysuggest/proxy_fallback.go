package shopifysuggest

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
)

// suggestAttemptTimeout bounds a single suggest endpoint attempt so a slow or
// throttling upstream cannot stall the whole fallback chain.
const suggestAttemptTimeout = 10 * time.Second

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
// transports, returning the first successful result. Each attempt only runs
// when the previous one errors, so a healthy direct connection never incurs
// proxy cost while a rate-limited endpoint still resolves via dedicated and
// then dynamic proxies.
func searchWithProxyFallback(ctx context.Context, cfg Config, apiURL string, mapProduct func(cfg Config, product Product) (gateway.Card, bool)) ([]gateway.Card, error) {
	return runSearchAttempts(ctx, buildSearchAttempts(), cfg, apiURL, mapProduct)
}

// runSearchAttempts executes the ordered attempts, returning the first result
// that succeeds (a nil error, even with zero cards). Each attempt's error is
// annotated with its position and strategy name so the final error reflects the
// last transport tried.
func runSearchAttempts(ctx context.Context, attempts []searchAttempt, cfg Config, apiURL string, mapProduct func(cfg Config, product Product) (gateway.Card, bool)) ([]gateway.Card, error) {
	var (
		cards []gateway.Card
		err   error
	)
	for idx, attempt := range attempts {
		cards, err = fetchAndMapProducts(ctx, attempt.client, apiURL, cfg, mapProduct)
		err = annotateSuggestAttemptError(idx+1, attempt.strategy, err)
		if err == nil {
			return cards, nil
		}
	}
	return cards, err
}

// buildSearchAttempts builds the ordered fallback chain. The direct attempt is
// always present; the dedicated and dynamic proxy attempts are appended only
// when their respective proxies are configured.
func buildSearchAttempts() []searchAttempt {
	attempts := []searchAttempt{
		{
			strategy: "direct",
			client:   &http.Client{Timeout: suggestAttemptTimeout},
		},
	}

	if client, ok := dedicatedProxyClient(); ok {
		attempts = append(attempts, searchAttempt{strategy: "dedicated", client: client})
	}

	if client, ok := dynamicProxyClient(); ok {
		attempts = append(attempts, searchAttempt{strategy: "dynamic", client: client})
	}

	return attempts
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
		Timeout: suggestAttemptTimeout,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsedProxyURL),
		},
	}, nil
}

func fetchAndMapProducts(ctx context.Context, client *http.Client, apiURL string, cfg Config, mapProduct func(cfg Config, product Product) (gateway.Card, bool)) ([]gateway.Card, error) {
	products, err := fetchProducts(ctx, client, apiURL)
	if err != nil {
		return nil, err
	}
	return mapProducts(cfg, products, mapProduct), nil
}

func annotateSuggestAttemptError(attempt int, strategy string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("attempt %d (%s): %w", attempt, strategy, err)
}
