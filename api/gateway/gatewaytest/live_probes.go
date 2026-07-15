package gatewaytest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// RequireAgoraSearchStructure verifies the Agora MTG search page markup.
func RequireAgoraSearchStructure(t *testing.T, ctx context.Context, baseURL, searchPath, category, query string) {
	t.Helper()
	probeURL := BuildURL("https", strings.TrimPrefix(strings.TrimPrefix(baseURL, "https://"), "http://"), searchPath, url.Values{
		"category":    {category},
		"searchfield": {query},
	})
	pageURL, err := url.Parse(probeURL)
	require.NoError(t, err)
	RequireHTMLStructure(t, ctx, HTMLProbe{
		URL:              probeURL,
		PrimarySelector:  "div#store_listingcontainer",
		FallbackSelector: "div.store-item",
		PageURL:          pageURL,
	})
}

// RequireFiveManaSearchStructure verifies the 5 Mana Shopify search page markup.
func RequireFiveManaSearchStructure(t *testing.T, ctx context.Context, baseURL, searchPath, query string) {
	t.Helper()
	host := strings.TrimPrefix(strings.TrimPrefix(baseURL, "https://"), "http://")
	probeURL := BuildURL("https", host, searchPath, url.Values{
		"q":                     {query},
		"filter.v.availability": {"1"},
		"section_id":            {"main-search"},
	})
	pageURL, err := url.Parse(probeURL)
	require.NoError(t, err)
	RequireHTMLStructure(t, ctx, HTMLProbe{
		URL:                    probeURL,
		PrimarySelector:        "ul.product-grid li",
		FallbackSelector:       "ul.product-grid",
		PageURL:                pageURL,
		ShopifySGDCurrency:     true,
		PreferResidentialProxy: true,
		SkipDirect:             true,
	})
}

// RequireDuellersPointSearchStructure verifies the Dueller's Point search table markup.
func RequireDuellersPointSearchStructure(t *testing.T, ctx context.Context, baseURL, searchPath, query string) {
	t.Helper()
	host := strings.TrimPrefix(strings.TrimPrefix(baseURL, "https://"), "http://")
	probeURL := BuildURL("https", host, searchPath, url.Values{"search_text": {query}})
	pageURL, err := url.Parse(probeURL)
	require.NoError(t, err)
	RequireHTMLStructure(t, ctx, HTMLProbe{
		URL:              probeURL,
		PrimarySelector:  "div.container table > tbody",
		FallbackSelector: "div.container table",
		PageURL:          pageURL,
	})
}

// RequireMoxAndLotusAPIStructure verifies the Mox & Lotus products API response shape.
func RequireMoxAndLotusAPIStructure(t *testing.T, ctx context.Context, query string) {
	t.Helper()
	probeURL := BuildURL("https", "moxandlotus.sg", "/api/products", url.Values{
		"limit":          {"48"},
		"full_search":    {"true"},
		"showStatus":     {"false"},
		"is_paginated":   {"true"},
		"in_stock":       {"true"},
		"sortVariation":  {"true"},
		"category_id":    {"1"},
		"variation_code": {"all"},
		"order_by":       {"Price Low to High"},
		"search":         {query},
	})
	RequireJSONStructure(t, ctx, JSONProbe{
		URL: probeURL,
		Validate: func(body []byte) error {
			var payload struct {
				Data []json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(body, &payload); err != nil {
				return err
			}
			if payload.Data == nil {
				return ValidateErrorf("missing data array")
			}
			return nil
		},
	})
}

// RequireCardsAndCollectionAPIStructure verifies the catalog search API response shape.
func RequireCardsAndCollectionAPIStructure(t *testing.T, ctx context.Context, baseURL, query string) {
	t.Helper()
	requestBody := fmt.Appendf(nil, `{"query":{"bool":{"should":[{"simple_query_string":{"query":"%s","fields":["name","setCode","setName"],"default_operator":"AND"}},{"multi_match":{"query":"%s","type":"phrase_prefix","fields":["name","setCode","setName"]}}]}},"post_filter":{"bool":{"must":{"terms":{"collectableContext.raw":["MTG","ACCESSORY"]}}}},"aggs":{"productCategory4":{"filter":{"bool":{"must":{"terms":{"collectableContext.raw":["MTG","ACCESSORY"]}}}},"aggs":{"productCategory.raw":{"terms":{"field":"productCategory.raw","size":50}},"productCategory.raw_count":{"cardinality":{"field":"productCategory.raw"}}}}},"size":20,"sort":[{"quantityOnSale":"desc"}]}`, query, query)
	RequireJSONStructure(t, ctx, JSONProbe{
		Method: "POST",
		URL:    strings.TrimRight(baseURL, "/") + "/api/catalog/",
		Body:   requestBody,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Validate: func(body []byte) error {
			var payload map[string]json.RawMessage
			if err := json.Unmarshal(body, &payload); err != nil {
				return err
			}
			if _, ok := payload["hits"]; !ok {
				return ValidateErrorf("missing hits key")
			}
			return nil
		},
	})
}

// RequireCardsCentralAPIStructure verifies the Cards Central LGS search API response shape.
func RequireCardsCentralAPIStructure(t *testing.T, ctx context.Context, baseURL, query string) {
	t.Helper()
	host := strings.TrimPrefix(strings.TrimPrefix(baseURL, "https://"), "http://")
	probeURL := BuildURL("https", host, "/api/lgs/search", url.Values{"q": {query}})
	RequireJSONStructure(t, ctx, JSONProbe{
		URL: probeURL,
		Headers: map[string]string{
			"Accept": "application/json",
		},
		Validate: func(body []byte) error {
			var payload []json.RawMessage
			if err := json.Unmarshal(body, &payload); err != nil {
				return err
			}
			return nil
		},
	})
}

// RequireTCGMarketplaceAPIStructure verifies the TCG Marketplace advanced search API shape.
func RequireTCGMarketplaceAPIStructure(t *testing.T, ctx context.Context, accessToken, query string) {
	t.Helper()
	requestBody := fmt.Appendf(nil, `{"access_token":%q,"name":%q,"category":3,"order":"name_asc"}`, accessToken, query)
	RequireJSONStructure(t, ctx, JSONProbe{
		Method: "POST",
		URL:    "https://thetcgmarketplace.com:3501/encoder/advancedsearch",
		Body:   requestBody,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Validate: func(body []byte) error {
			var payload map[string]json.RawMessage
			if err := json.Unmarshal(body, &payload); err != nil {
				return err
			}
			for _, key := range []string{"status", "data"} {
				if _, ok := payload[key]; !ok {
					return ValidateErrorf("missing %q key", key)
				}
			}
			return nil
		},
	})
}
