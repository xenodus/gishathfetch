package binderpos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
)

const storefrontSuggestPath = "/search/suggest.json"

type storefrontSuggestResponse struct {
	Resources struct {
		Results struct {
			Products []storefrontProduct `json:"products"`
		} `json:"results"`
	} `json:"resources"`
}

type storefrontProduct struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Image string `json:"image"`
}

type storefrontProductDetail struct {
	Title    string                   `json:"title"`
	Variants []storefrontProductStock `json:"variants"`
}

type storefrontProductStock struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Available bool   `json:"available"`
	Price     int    `json:"price"`
}

func (i impl) Search(ctx context.Context, scrapVariant int, storeName, baseURL, searchURL, searchStr string) ([]gateway.Card, error) {
	if !config.UseBinderposStorefrontAPI() {
		return i.Scrap(ctx, scrapVariant, storeName, baseURL, searchURL, searchStr)
	}

	cards, err := searchByStorefrontAPI(ctx, scrapVariant, storeName, baseURL, searchStr)
	if err != nil || len(cards) == 0 {
		return i.Scrap(ctx, scrapVariant, storeName, baseURL, searchURL, searchStr)
	}

	return cards, nil
}

func searchByStorefrontAPI(ctx context.Context, scrapVariant int, storeName, baseURL, searchStr string) ([]gateway.Card, error) {
	client := &http.Client{Timeout: config.PerSiteTimeout}

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
	query.Set("resources[limit]", "20")
	suggestURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, suggestURL.String(), nil)
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
		return nil, fmt.Errorf("storefront suggest request returned status %d", res.StatusCode)
	}

	var payload storefrontSuggestResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return payload.Resources.Results.Products, nil
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

func formatCardName(scrapVariant int, productTitle, variantTitle string) string {
	productTitle = strings.TrimSpace(productTitle)
	variantTitle = strings.TrimSpace(variantTitle)

	switch scrapVariant {
	case 2:
		if variantTitle == "" {
			return productTitle
		}
		return strings.TrimSpace(productTitle + " - " + variantTitle)
	case 3:
		return stripTrailingSet(productTitle)
	default:
		return productTitle
	}
}

func stripTrailingSet(productTitle string) string {
	title := strings.TrimSpace(productTitle)
	open := strings.LastIndex(title, "[")
	close := strings.LastIndex(title, "]")
	if open >= 0 && close > open && close == len(title)-1 {
		return strings.TrimSpace(title[:open])
	}
	return title
}

func extractSetName(productTitle string) string {
	title := strings.TrimSpace(productTitle)
	open := strings.LastIndex(title, "[")
	close := strings.LastIndex(title, "]")
	if open >= 0 && close > open && close == len(title)-1 {
		return strings.TrimSpace(title[open+1 : close])
	}
	return ""
}

func buildCardImageURL(rawImageURL, cardTitle string) string {
	img := strings.TrimSpace(rawImageURL)
	if strings.HasPrefix(img, "//") {
		return "https:" + img
	}
	if strings.HasPrefix(img, "http://") || strings.HasPrefix(img, "https://") {
		return img
	}
	return fmt.Sprintf("https://placehold.co/304x424?text=%s", url.QueryEscape(strings.TrimSpace(cardTitle)))
}
