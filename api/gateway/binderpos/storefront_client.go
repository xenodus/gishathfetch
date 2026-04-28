package binderpos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
)

func searchByStorefrontAPI(ctx context.Context, scrapVariant int, storeName, baseURL, shopifyDomain, searchStr string) ([]gateway.Card, error) {
	proxyURLs := util.GetDedicatedProxyURLs()
	if len(proxyURLs) == 0 {
		return nil, fmt.Errorf("no dedicated proxy configured for binderpos storefront api")
	}
	leasedURL, release, err := gateway.LeaseDedicatedProxyURL(ctx, proxyURLs)
	if err != nil {
		return nil, fmt.Errorf("dedicated proxy lease for binderpos storefront api: %w", err)
	}
	defer release()

	client, err := newHTTPClientWithProxyURL(leasedURL)
	if err != nil {
		return nil, fmt.Errorf("invalid dedicated proxy configured for binderpos storefront api: %w", err)
	}

	return searchByStorefrontAPIWithClient(ctx, client, scrapVariant, storeName, baseURL, shopifyDomain, searchStr)
}

func searchByStorefrontAPIDirect(ctx context.Context, scrapVariant int, storeName, baseURL, shopifyDomain, searchStr string) ([]gateway.Card, error) {
	client := &http.Client{Timeout: binderposAttemptTimeout}
	return searchByStorefrontAPIWithClient(ctx, client, scrapVariant, storeName, baseURL, shopifyDomain, searchStr)
}

func searchByStorefrontAPISharedProxy(ctx context.Context, scrapVariant int, storeName, baseURL, shopifyDomain, searchStr string) ([]gateway.Card, error) {
	sharedProxyURL := strings.TrimSpace(os.Getenv("PROXY_URL"))
	if sharedProxyURL == "" {
		return nil, fmt.Errorf("no shared proxy configured for binderpos storefront api")
	}

	client, err := newHTTPClientWithProxyURL(sharedProxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid shared proxy configured for binderpos storefront api: %w", err)
	}

	return searchByStorefrontAPIWithClient(ctx, client, scrapVariant, storeName, baseURL, shopifyDomain, searchStr)
}

func searchByStorefrontAPIWithClient(ctx context.Context, client *http.Client, scrapVariant int, storeName, baseURL, shopifyDomain, searchStr string) ([]gateway.Card, error) {
	if shouldUseDecklistEndpoint() {
		return searchByBinderposDecklistAPI(ctx, client, scrapVariant, storeName, baseURL, shopifyDomain, searchStr)
	}

	return searchByStorefrontProductDetailsAPI(ctx, client, scrapVariant, storeName, baseURL, searchStr)
}

func useDecklistForRoll(roll int) bool {
	return roll < binderposDecklistPct
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
