package binderpos

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
)

// binderposDedicatedProxySeq advances for each BinderPOS storefront API call that uses
// non-leased dedicated routing, so traffic round-robins across configured dedicated proxies.
var binderposDedicatedProxySeq atomic.Uint32

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
	if pinned, ok := gateway.RequestDedicatedProxyURL(ctx); ok {
		proxyURL = pinned
	} else if config.UseLeasedDedicatedProxy {
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

	return searchByBinderposDecklistAPI(ctx, client, scrapVariant, storeName, baseURL, shopifyDomain, searchStr)
}

func searchByStorefrontAPIDynamic(ctx context.Context, scrapVariant int, storeName, baseURL, shopifyDomain, searchStr string) ([]gateway.Card, error) {
	proxyURL := gateway.DynamicProxyURL()
	if proxyURL == "" {
		return nil, fmt.Errorf("no dynamic proxy configured for binderpos storefront api")
	}

	releaseDynamicProxy, err := gateway.AcquireDynamicProxySlot(ctx)
	if err != nil {
		return nil, err
	}
	defer releaseDynamicProxy()

	client, err := newHTTPClientWithProxyURL(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid dynamic proxy configured for binderpos storefront api: %w", err)
	}

	return searchByBinderposDecklistAPI(ctx, client, scrapVariant, storeName, baseURL, shopifyDomain, searchStr)
}

func searchByStorefrontAPIDirect(ctx context.Context, scrapVariant int, storeName, baseURL, shopifyDomain, searchStr string) ([]gateway.Card, error) {
	profile := gateway.PickBrowserProfile()
	if !gateway.ShouldUseBrowserTLSEmulationForScraping() {
		profile = gateway.BrowserEmulationProfile{}
	}
	client, err := gateway.NewBinderposHTTPClient("", profile)
	if err != nil {
		return nil, fmt.Errorf("binderpos direct client: %w", err)
	}
	return searchByBinderposDecklistAPI(ctx, client, scrapVariant, storeName, baseURL, shopifyDomain, searchStr)
}

func newHTTPClientWithProxyURL(proxyURL string) (*http.Client, error) {
	profile := gateway.PickBrowserProfile()
	if !gateway.ShouldUseBrowserTLSEmulationForScraping() {
		profile = gateway.BrowserEmulationProfile{}
	}
	return gateway.NewBinderposHTTPClient(proxyURL, profile)
}
