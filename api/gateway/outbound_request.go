package gateway

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// OutboundRequestStyle selects which browser-like headers to apply before signing.
type OutboundRequestStyle int

const (
	OutboundStyleNone OutboundRequestStyle = iota
	OutboundStyleHTML
	OutboundStyleJSON
)

// OutboundRequestOptions configures shared outbound gateway request preparation.
type OutboundRequestOptions struct {
	Style     OutboundRequestStyle
	PageURL   *url.URL
	StoreBase *url.URL
	// Accept sets an optional Accept header after style-specific headers apply.
	Accept string
	// SkipDirect omits the direct transport from DoOutboundGET fallback chains.
	SkipDirect bool
	// DirectAttemptTimeout overrides config.DirectSearchAttemptTimeout when > 0.
	DirectAttemptTimeout time.Duration
	// DedicatedAttemptTimeout overrides config.DedicatedSearchAttemptTimeout when > 0.
	DedicatedAttemptTimeout time.Duration
	// PreferDedicatedFirst is deprecated and ignored; outbound fallback always tries
	// direct before dedicated when both are enabled.
	PreferDedicatedFirst bool
	// PreferResidentialProxy prepends RESIDENTIAL_PROXY_1 to DoOutboundGET fallback
	// chains before dedicated transports.
	PreferResidentialProxy bool
	// OnlyProxyURL restricts DoOutboundGET to a single proxy transport when set.
	// Direct and dedicated fallbacks are skipped.
	OnlyProxyURL string
	// SkipWebBotAuth uses a browser User-Agent and omits Web Bot Auth signing.
	SkipWebBotAuth bool
	// ShopifySGDCurrency sets cart_currency/localization cookies for Shopify storefronts.
	ShopifySGDCurrency bool
	// ExtraHeaders sets additional request headers after style-specific headers apply.
	ExtraHeaders map[string]string
}

// PrepareOutboundRequest applies per-domain pacing, browser-like headers, a
// consistent User-Agent, and optional Web Bot Auth signing to an outbound request.
func PrepareOutboundRequest(ctx context.Context, req *http.Request, opts OutboundRequestOptions) error {
	if req == nil || req.URL == nil {
		return nil
	}
	if err := WaitForDomainRequestSlot(ctx, req.URL); err != nil {
		return err
	}

	switch opts.Style {
	case OutboundStyleHTML:
		ApplyBrowserLikeHTMLHeaders(&req.Header, opts.PageURL)
		req.Header.Set("Accept-Encoding", "gzip")
	case OutboundStyleJSON:
		storeBase := opts.StoreBase
		if storeBase == nil {
			storeBase = opts.PageURL
		}
		ApplyBrowserLikeJSONFetchHeaders(&req.Header, storeBase)
	}

	if opts.Accept != "" {
		req.Header.Set("Accept", opts.Accept)
	}

	if opts.ShopifySGDCurrency {
		ApplyShopifySGDCurrencyCookie(&req.Header)
	}

	for key, value := range opts.ExtraHeaders {
		if key != "" && value != "" {
			req.Header.Set(key, value)
		}
	}

	profile := ResolveBrowserProfileForOutbound(ctx, opts)
	if opts.SkipWebBotAuth {
		if profile.Enabled {
			req.Header.Set("User-Agent", profile.UserAgent)
			ApplyBrowserProfileHeaders(&req.Header, profile)
		} else {
			req.Header.Set("User-Agent", RandomBrowserUserAgent())
		}
		return nil
	}

	if profile.Enabled {
		req.Header.Set("User-Agent", profile.UserAgent)
		ApplyBrowserProfileHeaders(&req.Header, profile)
		return nil
	}

	req.Header.Set("User-Agent", OutboundUserAgent())
	return SignWebBotAuthRequest(req)
}
