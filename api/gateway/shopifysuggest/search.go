package shopifysuggest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"mtg-price-checker-sg/gateway"
)

const SearchPath = "/search/suggest.json"

// predictiveSearchLimit is the maximum number of products Shopify's predictive
// search endpoint will return per request (the platform caps this at 10).
const predictiveSearchLimit = "10"

// Config identifies a Shopify storefront that exposes predictive search.
type Config struct {
	StoreName string
	BaseURL   string
}

// Options controls query shaping and product-to-card mapping for a search.
type Options struct {
	Config
	SearchStr string
	// BuildQuery transforms the raw search string into the Shopify q= parameter.
	// When nil, SearchStr is passed through unchanged.
	BuildQuery func(searchStr string) string
	// QueryValues builds the full suggest endpoint query. When nil, binderpos-style
	// defaults are used (unavailable products last, predictive fields).
	QueryValues func(searchStr string) url.Values
	// MapProduct converts one suggest product into a card. Return false to skip.
	MapProduct func(cfg Config, product Product) (gateway.Card, bool)
}

// Product is one item from Shopify's /search/suggest.json response.
type Product struct {
	Title         string   `json:"title"`
	Handle        string   `json:"handle"`
	URL           string   `json:"url"`
	Price         string   `json:"price"`
	Available     bool     `json:"available"`
	Image         string   `json:"image"`
	ProductType   string   `json:"type"`
	Vendor        string   `json:"vendor"`
	Tags          []string `json:"tags"`
	FeaturedImage struct {
		URL string `json:"url"`
	} `json:"featured_image"`
}

type suggestResponse struct {
	Resources struct {
		Results struct {
			Products []Product `json:"products"`
		} `json:"results"`
	} `json:"resources"`
}

// Search queries a Shopify storefront's predictive search endpoint and maps
// products into cards using the supplied options.
//
// Shopify's predictive search endpoint can rate limit (HTTP 429) or otherwise
// throttle direct traffic, so the request is attempted through an ordered
// fallback chain: a direct connection first, then a dedicated proxy, and
// finally the shared dynamic proxy. The first attempt that succeeds wins;
// later attempts only run when the previous one errors (network failure,
// non-200 status such as 429, or an unparsable body).
func Search(ctx context.Context, opts Options) ([]gateway.Card, error) {
	mapProduct := opts.MapProduct
	if mapProduct == nil {
		return nil, fmt.Errorf("shopifysuggest: MapProduct is required")
	}

	apiURL, err := buildSuggestURL(opts)
	if err != nil {
		return nil, err
	}

	return searchWithProxyFallback(ctx, opts.Config, apiURL, mapProduct)
}

// buildSuggestURL assembles the full predictive search endpoint URL from the
// store config and query-shaping options.
func buildSuggestURL(opts Options) (string, error) {
	host, err := hostFromBaseURL(opts.Config.BaseURL)
	if err != nil {
		return "", err
	}

	searchQuery := strings.TrimSpace(opts.SearchStr)
	if opts.BuildQuery != nil {
		searchQuery = strings.TrimSpace(opts.BuildQuery(opts.SearchStr))
	}

	queryValues := opts.QueryValues
	if queryValues == nil {
		queryValues = BinderposQueryValues
	}
	values := queryValues(searchQuery)

	apiURL := url.URL{
		Scheme:   "https",
		Host:     host,
		Path:     SearchPath,
		RawQuery: values.Encode(),
	}
	return apiURL.String(), nil
}

// fetchProducts performs a single suggest request with the supplied client and
// returns the raw products. A non-200 status (e.g. 429 Too Many Requests) is
// reported as an error so the caller can fall back to another transport.
func fetchProducts(ctx context.Context, client *http.Client, apiURL string) ([]Product, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", gateway.RandomBrowserUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shopifysuggest: unexpected status %d", resp.StatusCode)
	}

	var res suggestResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	return res.Resources.Results.Products, nil
}

// mapProducts filters the suggest products down to in-stock Magic cards and
// maps them into gateway cards using the supplied mapper.
func mapProducts(cfg Config, products []Product, mapProduct func(cfg Config, product Product) (gateway.Card, bool)) []gateway.Card {
	cards := make([]gateway.Card, 0, len(products))
	for _, product := range products {
		if !product.Available {
			continue
		}
		if !IsMagicProduct(product.ProductType, product.Vendor, product.Tags) {
			continue
		}

		card, ok := mapProduct(cfg, product)
		if !ok {
			continue
		}
		cards = append(cards, card)
	}
	return cards
}

func hostFromBaseURL(baseURL string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", fmt.Errorf("invalid base URL %q: %w", baseURL, err)
	}
	if u.Host == "" {
		return "", fmt.Errorf("invalid base URL %q: missing host", baseURL)
	}
	return u.Host, nil
}
