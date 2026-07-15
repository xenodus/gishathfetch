package gatewaytest

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"mtg-price-checker-sg/gateway"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"
)

// HTMLProbe fetches a storefront page and checks that expected markup is present.
type HTMLProbe struct {
	URL                    string
	PrimarySelector        string
	FallbackSelector       string
	PageURL                *url.URL
	ShopifySGDCurrency     bool
	PreferResidentialProxy bool
	SkipDirect             bool
	SkipWebBotAuth         bool
}

// RequireHTMLStructure verifies the probe URL returns HTTP 200 and contains the
// expected selectors. Inventory may be empty; the check only guards page shape.
func RequireHTMLStructure(t *testing.T, ctx context.Context, probe HTMLProbe) {
	t.Helper()
	require.NotEmpty(t, probe.URL)
	require.NotEmpty(t, probe.PrimarySelector)

	pageURL := probe.PageURL
	if pageURL == nil {
		parsed, err := url.Parse(probe.URL)
		require.NoError(t, err)
		pageURL = parsed
	}

	resp, err := gateway.DoOutboundGET(ctx, probe.URL, gateway.OutboundRequestOptions{
		Style:                  gateway.OutboundStyleHTML,
		PageURL:                pageURL,
		ShopifySGDCurrency:     probe.ShopifySGDCurrency,
		PreferResidentialProxy:   probe.PreferResidentialProxy,
		SkipDirect:               probe.SkipDirect,
		SkipWebBotAuth:           probe.SkipWebBotAuth,
	}, 20*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "expected HTML page %s to return 200", probe.URL)

	body, err := gateway.ReadResponseBody(resp)
	require.NoError(t, err)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	require.NoError(t, err)

	if doc.Find(probe.PrimarySelector).Length() > 0 {
		return
	}
	if probe.FallbackSelector != "" && doc.Find(probe.FallbackSelector).Length() > 0 {
		return
	}

	require.Failf(t, "expected HTML structure not found",
		"page %s is missing selectors %q (fallback %q)", probe.URL, probe.PrimarySelector, probe.FallbackSelector)
}

// RequireHTMLBodyContains verifies the fetched page body includes a marker string.
func RequireHTMLBodyContains(t *testing.T, ctx context.Context, probeURL string, pageURL *url.URL, marker string) {
	t.Helper()
	require.NotEmpty(t, marker)

	resp, err := gateway.DoOutboundGET(ctx, probeURL, gateway.OutboundRequestOptions{
		Style:   gateway.OutboundStyleHTML,
		PageURL: pageURL,
	}, 20*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := gateway.ReadResponseBody(resp)
	require.NoError(t, err)
	require.Contains(t, string(body), marker, fmt.Sprintf("page %s missing marker %q", probeURL, marker))
}
