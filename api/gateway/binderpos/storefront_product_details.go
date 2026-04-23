package binderpos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
)

func searchByStorefrontProductDetailsAPI(ctx context.Context, client *http.Client, scrapVariant int, storeName, baseURL, searchStr string) ([]gateway.Card, error) {
	products, err := fetchSuggestProducts(ctx, client, baseURL, searchStr)
	if err != nil {
		return nil, err
	}
	if len(products) == 0 {
		return []gateway.Card{}, nil
	}

	cards := make([]gateway.Card, 0, len(products)*4)
	seen := map[string]struct{}{}
	for _, product := range products {
		productURL := product.URL
		if productURL == "" {
			continue
		}

		detail, err := fetchProductDetail(ctx, client, baseURL, productURL)
		if err != nil {
			continue
		}

		for _, variant := range detail.Variants {
			if variant.Price <= 0 {
				continue
			}

			quality := strings.TrimSpace(strings.ReplaceAll(variant.Title, "Foil", ""))
			cardURL, err := buildProductURLWithVariant(baseURL, productURL, variant.ID)
			if err != nil {
				continue
			}

			setName := extractSetName(detail.Title)
			image := buildCardImageURL(product.Image, detail.Title)
			card := gateway.Card{
				Name:    formatCardName(scrapVariant, detail.Title, variant.Title),
				Url:     cardURL,
				Img:     image,
				Price:   float64(variant.Price) / 100,
				InStock: variant.Available,
				IsFoil:  strings.Contains(strings.ToLower(variant.Title), "foil"),
				Source:  storeName,
				Quality: util.MapQuality(quality),
			}
			if scrapVariant == 3 && setName != "" {
				card.ExtraInfo = []string{setName}
			}

			key := fmt.Sprintf("%s|%.2f|%t", card.Url, card.Price, card.InStock)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			cards = append(cards, card)
		}
	}

	return cards, nil
}

func fetchSuggestProducts(ctx context.Context, client *http.Client, baseURL, searchStr string) ([]storefrontProduct, error) {
	suggestURL, err := url.Parse(baseURL + storefrontSuggestPath)
	if err != nil {
		return nil, err
	}

	query := suggestURL.Query()
	query.Set("q", strings.TrimSpace(searchStr+" mtg"))
	query.Set("resources[type]", "product")
	query.Set("resources[limit]", "8")
	query.Set("resources[options][unavailable_products]", "hide")
	suggestURL.RawQuery = query.Encode()

	var payload storefrontSuggestResponse
	fullURL := suggestURL.String()
	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", gateway.RandomBrowserUserAgent())
		if err := gateway.WaitForDomainRequestSlot(ctx, req.URL); err != nil {
			return nil, err
		}

		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode == http.StatusOK {
			if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
				res.Body.Close()
				return nil, err
			}
			res.Body.Close()
			return payload.Resources.Results.Products, nil
		}
		_, _ = io.Copy(io.Discard, res.Body)
		res.Body.Close()
		if attempt < 3 && isRetriableHTTPStatus(res.StatusCode) {
			time.Sleep(400 * time.Duration(attempt) * time.Millisecond)
			continue
		}
		return nil, fmt.Errorf("storefront suggest request returned status %d", res.StatusCode)
	}
	// 3 fixed iterations always return from inside the loop; keep for the compiler and linters.
	return nil, fmt.Errorf("storefront suggest: exhausted transient retries")
}

func isRetriableHTTPStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func fetchProductDetail(ctx context.Context, client *http.Client, baseURL, productPath string) (*storefrontProductDetail, error) {
	productJSONURL, err := productPathToJSONURL(baseURL, productPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, productJSONURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", gateway.RandomBrowserUserAgent())
	if err := gateway.WaitForDomainRequestSlot(ctx, req.URL); err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("product detail request failed status=%d body=%s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	var detail storefrontProductDetail
	if err := json.NewDecoder(res.Body).Decode(&detail); err != nil {
		return nil, err
	}

	return &detail, nil
}

func productPathToJSONURL(baseURL, productPath string) (string, error) {
	u, err := url.Parse(productPath)
	if err != nil {
		return "", err
	}

	path := strings.TrimSpace(u.Path)
	if path == "" {
		return "", fmt.Errorf("missing product path")
	}

	path = strings.TrimSuffix(path, "/")
	if !strings.HasSuffix(path, ".js") {
		path += ".js"
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	base.Path = path
	base.RawQuery = ""

	return base.String(), nil
}

func buildProductURLWithVariant(baseURL, productPath string, variantID int64) (string, error) {
	u, err := url.Parse(baseURL + productPath)
	if err != nil {
		return "", err
	}

	u.RawQuery = ""
	query := u.Query()
	query.Set("variant", fmt.Sprint(variantID))
	query.Set("utm_source", config.UtmSource)
	u.RawQuery = query.Encode()

	return u.String(), nil
}
