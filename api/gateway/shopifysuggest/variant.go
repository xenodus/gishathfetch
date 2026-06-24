package shopifysuggest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
)

// productJSONPathFormat builds the path to a Shopify product's JSON document
// from its handle. Unlike predictive search, this document exposes per-variant
// price and stock.
const productJSONPathFormat = "/products/%s.js"

// productDetail is the subset of Shopify's /products/<handle>.js document used
// to recover the per-variant price and availability that predictive search
// omits.
type productDetail struct {
	Variants []productVariant `json:"variants"`
}

type productVariant struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	// Price is reported in the store's minor currency unit (cents).
	Price     int64 `json:"price"`
	Available bool  `json:"available"`
}

// resolveVariantCards expands a single suggest product into one card per
// in-stock variant, sourcing the real price and condition from the product's
// JSON document. The base card supplies the fields that are constant across a
// product's variants (name, set, image, URL); the variant supplies price,
// stock, condition, and foil.
//
// If the detail request fails, the predictive-search card is returned as a
// resilience fallback so a transient upstream error does not drop the store
// from results. That fallback price is the predictive-search price_min, which
// is the pre-existing (pre-resolution) behavior.
func resolveVariantCards(ctx context.Context, client *http.Client, cfg Config, product Product, base gateway.Card) []gateway.Card {
	detail, err := fetchProductDetail(ctx, client, cfg.BaseURL, product.Handle)
	if err != nil {
		log.Printf("shopifysuggest: variant resolution failed for %s [%s]: %v", cfg.StoreName, product.Handle, err)
		return []gateway.Card{base}
	}

	cards := make([]gateway.Card, 0, len(detail.Variants))
	for _, variant := range detail.Variants {
		if !variant.Available || variant.Price <= 0 {
			continue
		}

		card := base
		card.Price = float64(variant.Price) / 100
		card.InStock = true
		card.IsFoil = strings.Contains(strings.ToLower(variant.Title), "foil")
		card.Quality = util.MapQuality(strings.TrimSpace(strings.ReplaceAll(variant.Title, "Foil", "")))
		card.Url = withVariantParam(base.Url, variant.ID)
		cards = append(cards, card)
	}

	return cards
}

// fetchProductDetail retrieves and decodes a Shopify product JSON document for
// the given handle using the supplied client (which carries whichever transport
// won the suggest fallback).
func fetchProductDetail(ctx context.Context, client *http.Client, baseURL, handle string) (productDetail, error) {
	handle = strings.TrimSpace(handle)
	if handle == "" {
		return productDetail{}, fmt.Errorf("shopifysuggest: empty product handle")
	}

	base, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return productDetail{}, fmt.Errorf("shopifysuggest: invalid base URL %q: %w", baseURL, err)
	}
	if base.Host == "" {
		return productDetail{}, fmt.Errorf("shopifysuggest: invalid base URL %q: missing host", baseURL)
	}

	// Preserve the base URL's scheme and host; only the path identifies the
	// product JSON document.
	detailURL := *base
	detailURL.Path = fmt.Sprintf(productJSONPathFormat, handle)
	detailURL.RawQuery = ""
	detailURL.Fragment = ""

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, detailURL.String(), nil)
	if err != nil {
		return productDetail{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", gateway.RandomBrowserUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return productDetail{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return productDetail{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return productDetail{}, fmt.Errorf("shopifysuggest: unexpected status %d for %s", resp.StatusCode, detailURL.String())
	}

	var detail productDetail
	if err := json.Unmarshal(body, &detail); err != nil {
		return productDetail{}, err
	}
	return detail, nil
}

// withVariantParam adds a variant query parameter to an existing product URL,
// preserving any parameters already present (e.g. utm_source). On a parse
// failure it returns the original URL unchanged so a card is still produced.
func withVariantParam(rawURL string, variantID int64) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	query := u.Query()
	query.Set("variant", strconv.FormatInt(variantID, 10))
	u.RawQuery = query.Encode()
	return u.String()
}
