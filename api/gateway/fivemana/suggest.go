package fivemana

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
)

const storefrontSuggestPath = "/search/suggest.json"
const suggestProductLimit = "20"

type suggestResponse struct {
	Resources struct {
		Results struct {
			Products []suggestProduct `json:"products"`
		} `json:"results"`
	} `json:"resources"`
}

type suggestProduct struct {
	Available bool     `json:"available"`
	Handle    string   `json:"handle"`
	Image     string   `json:"image"`
	Price     string   `json:"price"`
	Tags      []string `json:"tags"`
	Title     string   `json:"title"`
	URL       string   `json:"url"`
}

func (s Store) searchSuggest(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	storeBase, err := s.storeBaseURL()
	if err != nil {
		return nil, err
	}

	apiURL := &url.URL{
		Scheme: storeBase.Scheme,
		Host:   storeBase.Host,
		Path:   storefrontSuggestPath,
		RawQuery: url.Values{
			"q": {searchStr},
			"resources[type]":                          {"product"},
			"resources[limit]":                         {suggestProductLimit},
			"resources[options][unavailable_products]": {"hide"},
		}.Encode(),
	}

	opts := fiveManaOutboundOpts(storeBase, storeBase, gateway.OutboundStyleJSON)
	// Shopify storefronts behind Cloudflare respond better to browser-like
	// requests than signed bot traffic on the suggest API path.
	opts.SkipWebBotAuth = true

	resp, err := gateway.DoOutboundGET(ctx, apiURL.String(), opts, config.SearchAttemptTimeout)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := gateway.ReadResponseBody(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%s", gateway.FormatUnexpectedHTTPStatus(s.Name, resp, body))
	}

	var parsed suggestResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, gateway.WrapJSONDecodeError(err, resp, body)
	}

	return mapSuggestProducts(s.Name, parsed.Resources.Results.Products), nil
}

func mapSuggestProducts(storeName string, products []suggestProduct) []gateway.Card {
	cards := make([]gateway.Card, 0, len(products))
	for _, product := range products {
		card, ok := mapSuggestProduct(storeName, product)
		if ok {
			cards = append(cards, card)
		}
	}
	return cards
}

func mapSuggestProduct(storeName string, product suggestProduct) (gateway.Card, bool) {
	if !product.Available {
		return gateway.Card{}, false
	}

	title := strings.TrimSpace(product.Title)
	if title == "" {
		return gateway.Card{}, false
	}

	price, err := util.ParsePrice(product.Price)
	if err != nil || price <= 0 {
		return gateway.Card{}, false
	}

	cardURL, err := suggestProductURL(product)
	if err != nil {
		return gateway.Card{}, false
	}

	name, isFoil := parseNameAndFoil(title)
	if name == "" {
		return gateway.Card{}, false
	}
	if !isFoil {
		isFoil = suggestProductIsFoil(product.Tags)
	}

	return gateway.Card{
		Name:      name,
		Url:       cardURL,
		Img:       strings.TrimSpace(product.Image),
		Price:     price,
		InStock:   true,
		IsFoil:    isFoil,
		Source:    storeName,
		ExtraInfo: extraInfoFromTitle(title),
	}, true
}

func suggestProductURL(product suggestProduct) (string, error) {
	handle := strings.TrimSpace(product.Handle)
	if handle == "" {
		return "", fmt.Errorf("missing product handle")
	}
	return productURLWithUTM(StoreBaseURL + "/products/" + handle)
}

func suggestProductIsFoil(tags []string) bool {
	return tagsIndicateFoil(tags)
}
