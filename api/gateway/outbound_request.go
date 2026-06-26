package gateway

import (
	"context"
	"net/http"
	"net/url"
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

	req.Header.Set("User-Agent", OutboundUserAgent())
	return SignWebBotAuthRequest(req)
}
