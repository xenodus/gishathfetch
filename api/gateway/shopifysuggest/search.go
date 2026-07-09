package shopifysuggest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"mtg-price-checker-sg/gateway"
)

const SearchPath = "/search/suggest.json"

// predictiveSearchLimit is the maximum number of products Shopify's predictive
// search endpoint will return per request (the platform caps this at 10).
const predictiveSearchLimit = "10"

// ShopifyLocalizationSingapore is the Shopify market cookie value for
// Singapore storefronts. Without it, Accept-Language defaults can return
// prices from a different market than the public site shows to local users.
const ShopifyLocalizationSingapore = "SG"

// Config identifies a Shopify storefront that exposes predictive search.
type Config struct {
	StoreName string
	BaseURL   string
	// ShopifyLocalization, when set, is sent as Shopify's localization cookie on
	// suggest and product JSON requests so responses use that market's prices.
	ShopifyLocalization string
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
	// ResolveVariants, when true, follows each suggest product with a request to
	// its Shopify product JSON endpoint (/products/<handle>.js) to emit one card
	// per in-stock variant using the variant's real price and condition.
	//
	// Shopify's predictive search reports a product-level price equal to the
	// cheapest variant regardless of stock (price_min), so a product whose only
	// cheap variants are sold out surfaces a price that cannot be purchased.
	// Variant resolution reads per-variant price and availability (absent from
	// predictive search) so the cheapest *in-stock* price is reported instead.
	ResolveVariants bool
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
// throttle direct traffic. Each transport retries transient failures on the
// same connection, honoring Retry-After when present, before the request is
// attempted through an ordered fallback chain: a direct connection first, then
// a dedicated proxy, and finally the shared dynamic proxy. The first transport
// that succeeds wins; later transports only run when the previous one exhausts
// its retries (network failure, persistent non-200 status such as 429, or an
// unparsable body).
func Search(ctx context.Context, opts Options) ([]gateway.Card, error) {
	if opts.MapProduct == nil {
		return nil, fmt.Errorf("shopifysuggest: MapProduct is required")
	}

	apiURL, err := buildSuggestURL(opts)
	if err != nil {
		return nil, err
	}

	return searchWithProxyFallback(ctx, opts, apiURL)
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

// fetchProducts performs a suggest request with the supplied client and returns
// the raw products. Transient rate-limit/5xx responses are retried on the same
// transport, honoring Retry-After when Shopify provides it. Persistent failures
// are reported as errors so the caller can fall back to another transport.
func fetchProducts(ctx context.Context, client *http.Client, apiURL string, reqOpts suggestRequestOpts) ([]Product, error) {
	body, err := doSuggestGETWithRetry(ctx, client, apiURL, reqOpts)
	if err != nil {
		return nil, err
	}

	var res suggestResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	return res.Resources.Results.Products, nil
}

// mapProducts filters the suggest products down to in-stock Magic cards and
// maps them into gateway cards. When opts.ResolveVariants is set, each product
// is expanded into one card per in-stock variant (see resolveVariantCards);
// otherwise a single card is emitted from the predictive-search fields.
//
// The HTTP client is threaded through so variant resolution reuses the same
// transport that won the suggest fallback (direct, dedicated, or dynamic
// proxy), avoiding a fresh, possibly-throttled connection per product.
func mapProducts(ctx context.Context, client *http.Client, opts Options, products []Product) []gateway.Card {
	cards := make([]gateway.Card, 0, len(products))
	for _, product := range products {
		if !product.Available {
			continue
		}
		if !IsMagicProduct(product.ProductType, product.Vendor, product.Tags) {
			continue
		}

		base, ok := opts.MapProduct(opts.Config, product)
		if !ok {
			continue
		}

		if !opts.ResolveVariants {
			cards = append(cards, base)
			continue
		}

		cards = append(cards, resolveVariantCards(ctx, client, opts.Config, product, base)...)
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
