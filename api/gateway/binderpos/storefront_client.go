package binderpos

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

// binderposDedicatedProxySeq advances for each BinderPOS storefront API call that uses
// non-leased dedicated routing, so traffic round-robins across configured dedicated proxies.
var binderposDedicatedProxySeq atomic.Uint32

// binderposDecklistRouteSeq advances once per first-attempt storefront routing
// decision so that decklist vs. product-details selection round-robins evenly
// across the stores in a search instead of every store converging on the shared
// portal host.
var binderposDecklistRouteSeq atomic.Uint32

// nextBinderposStorefrontProxyURL returns the next dedicated proxy URL in round-robin order.
func nextBinderposStorefrontProxyURL(proxyURLs []string) string {
	n := len(proxyURLs)
	v := binderposDedicatedProxySeq.Add(1) - 1
	slot := int(v % uint32(n))
	return proxyURLs[slot]
}

func searchByStorefrontAPI(ctx context.Context, scrapVariant int, storeName, baseURL, shopifyDomain, searchStr string) ([]gateway.Card, error) {
	proxyURLs := util.GetDedicatedProxyURLs()
	if len(proxyURLs) == 0 {
		return nil, fmt.Errorf("no dedicated proxy configured for binderpos storefront api")
	}

	var proxyURL string
	if config.UseLeasedDedicatedProxy {
		leasedURL, release, err := gateway.LeaseDedicatedProxyURL(ctx, proxyURLs)
		if err != nil {
			return nil, fmt.Errorf("dedicated proxy lease for binderpos storefront api: %w", err)
		}
		defer release()
		proxyURL = leasedURL
	} else {
		proxyURL = nextBinderposStorefrontProxyURL(proxyURLs)
	}

	client, err := newHTTPClientWithProxyURL(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid dedicated proxy configured for binderpos storefront api: %w", err)
	}

	return searchByStorefrontAPIWithClient(ctx, client, scrapVariant, storeName, baseURL, shopifyDomain, searchStr)
}

func searchByStorefrontAPIDynamic(ctx context.Context, scrapVariant int, storeName, baseURL, shopifyDomain, searchStr string) ([]gateway.Card, error) {
	proxyURL := gateway.DynamicProxyURL()
	if proxyURL == "" {
		return nil, fmt.Errorf("no dynamic proxy configured for binderpos storefront api")
	}

	client, err := newHTTPClientWithProxyURL(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid dynamic proxy configured for binderpos storefront api: %w", err)
	}

	return searchByStorefrontAPIWithClient(ctx, client, scrapVariant, storeName, baseURL, shopifyDomain, searchStr)
}

func searchByStorefrontAPIDirect(ctx context.Context, scrapVariant int, storeName, baseURL, shopifyDomain, searchStr string) ([]gateway.Card, error) {
	client := &http.Client{Timeout: binderposAttemptTimeout}
	return searchByStorefrontAPIWithClient(ctx, client, scrapVariant, storeName, baseURL, shopifyDomain, searchStr)
}

func searchByStorefrontAPIWithClient(ctx context.Context, client *http.Client, scrapVariant int, storeName, baseURL, shopifyDomain, searchStr string) ([]gateway.Card, error) {
	if shouldUseDecklistEndpoint() {
		return searchByBinderposDecklistAPI(ctx, client, scrapVariant, storeName, baseURL, shopifyDomain, searchStr)
	}

	return searchByStorefrontProductDetailsAPI(ctx, client, scrapVariant, storeName, baseURL, searchStr)
}

// useDecklistForRoute returns true for even sequence numbers, yielding a 50/50
// split between the shared decklist portal and the per-store product-details
// path across consecutive routing decisions.
func useDecklistForRoute(seq uint32) bool {
	return seq%2 == 0
}

func newHTTPClientWithProxyURL(proxyURL string) (*http.Client, error) {
	parsedProxyURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout: binderposAttemptTimeout,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsedProxyURL),
		},
	}, nil
}
