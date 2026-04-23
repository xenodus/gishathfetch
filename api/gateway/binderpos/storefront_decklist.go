package binderpos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
)

func searchByBinderposDecklistAPI(ctx context.Context, client *http.Client, scrapVariant int, storeName, baseURL, searchStr string) ([]gateway.Card, error) {
	shopifyDomain, ok := storefrontShopifyDomainForBaseURL(baseURL)
	if !ok {
		return nil, fmt.Errorf("missing shopify domain mapping for binderpos storefront")
	}

	decklistURL, err := url.Parse(binderposDecklistAPIURL)
	if err != nil {
		return nil, err
	}
	query := decklistURL.Query()
	query.Set("storeUrl", shopifyDomain)
	query.Set("type", binderposDecklistType)
	decklistURL.RawQuery = query.Encode()

	payload := []storefrontDecklistRequestItem{
		{
			Card:     strings.TrimSpace(searchStr),
			Quantity: 1,
		},
	}
	payloadBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, decklistURL.String(), bytes.NewReader(payloadBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", gateway.RandomBrowserUserAgent())
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
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
		return nil, fmt.Errorf("binderpos decklist request failed status=%d body=%s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	var lines []storefrontDecklistLine
	if err := json.NewDecoder(res.Body).Decode(&lines); err != nil {
		return nil, err
	}

	return mapDecklistLinesToCards(scrapVariant, storeName, baseURL, lines), nil
}

func storefrontShopifyDomainForBaseURL(baseURL string) (string, bool) {
	parsedURL, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", false
	}

	host := strings.ToLower(strings.TrimSpace(parsedURL.Hostname()))
	host = strings.TrimPrefix(host, "www.")
	if host == "" {
		return "", false
	}

	shopifyDomain, ok := binderposShopifyDomainByStoreHost[host]
	return shopifyDomain, ok
}

func mapDecklistLinesToCards(scrapVariant int, storeName, baseURL string, lines []storefrontDecklistLine) []gateway.Card {
	cards := make([]gateway.Card, 0, len(lines)*4)
	seen := map[string]struct{}{}

	for _, line := range lines {
		for _, product := range line.Products {
			productTitle := strings.TrimSpace(product.Title)
			if productTitle == "" {
				productTitle = strings.TrimSpace(product.Name)
			}
			if productTitle == "" {
				productTitle = strings.TrimSpace(line.ValidName)
			}
			if productTitle == "" {
				continue
			}

			productPath := ""
			if strings.TrimSpace(product.Handle) != "" {
				productPath = "/products/" + strings.TrimSpace(product.Handle)
			}

			setName := strings.TrimSpace(product.SetName)
			if setName == "" {
				setName = extractSetName(productTitle)
			}

			image := buildCardImageURL(product.Image, productTitle)
			for _, variant := range product.Variants {
				if variant.Price <= 0 {
					continue
				}
				if variant.ShopifyID <= 0 || productPath == "" {
					continue
				}

				cardURL, err := buildProductURLWithVariant(baseURL, productPath, variant.ShopifyID)
				if err != nil {
					continue
				}

				quality := strings.TrimSpace(strings.ReplaceAll(variant.Title, "Foil", ""))
				card := gateway.Card{
					Name:    formatCardName(scrapVariant, productTitle, variant.Title),
					Url:     cardURL,
					Img:     image,
					Price:   variant.Price,
					InStock: variant.Quantity > 0,
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
	}

	return cards
}

