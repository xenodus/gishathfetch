package fivemana

import (
	"context"
	"encoding/json"
	"log"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
	"net/url"
	"strings"
)

const StoreName = "5 Mana"
const StoreBaseURL = "https://5-mana.sg"
const StoreSuggestPath = "/search/suggest.json"
const suggestProductLimit = "10"

type Store struct {
	Name        string
	BaseUrl     string
	SuggestPath string
}

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

func NewLGS() gateway.LGS {
	return Store{
		Name:        StoreName,
		BaseUrl:     StoreBaseURL,
		SuggestPath: StoreSuggestPath,
	}
}

func (s Store) Search(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	storeBase, err := url.Parse(StoreBaseURL)
	if err != nil {
		return nil, err
	}

	apiURL := &url.URL{
		Scheme: "https",
		Host:   "5-mana.sg",
		Path:   StoreSuggestPath,
		RawQuery: url.Values{
			"q": {searchStr},
			"resources[type]":                              {"product"},
			"resources[limit]":                             {suggestProductLimit},
			"resources[options][unavailable_products]":     {"hide"},
		}.Encode(),
	}

	resp, err := gateway.DoOutboundGET(ctx, apiURL.String(), gateway.OutboundRequestOptions{
		Style:              gateway.OutboundStyleJSON,
		StoreBase:          storeBase,
		ShopifySGDCurrency: true,
		SkipDirect:         true,
	}, config.SearchAttemptTimeout)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := gateway.ReadResponseBody(resp)
	if err != nil {
		return nil, err
	}

	var parsed suggestResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	return mapSuggestProductsToCards(parsed.Resources.Results.Products, s.Name), nil
}

func mapSuggestProductsToCards(products []suggestProduct, storeName string) []gateway.Card {
	cards := make([]gateway.Card, 0, len(products))
	for _, product := range products {
		card, ok := mapSuggestProductToCard(product, storeName)
		if ok {
			cards = append(cards, card)
		}
	}
	return cards
}

func mapSuggestProductToCard(product suggestProduct, storeName string) (gateway.Card, bool) {
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

	productURL, ok := buildSuggestProductURL(product)
	if !ok {
		return gateway.Card{}, false
	}

	c := gateway.Card{
		Name:    strings.TrimSpace(strings.Replace(title, "[Foil]", "", -1)),
		Url:     productURL,
		Img:     strings.TrimSpace(product.Image),
		Price:   price,
		InStock: true,
		IsFoil:  isSuggestProductFoil(title, product.Tags),
		Source:  storeName,
	}

	return c, c.Name != ""
}

func buildSuggestProductURL(product suggestProduct) (string, bool) {
	handle := strings.TrimSpace(product.Handle)
	if handle == "" {
		return "", false
	}

	cleanPageURL, err := url.Parse(StoreBaseURL + "/products/" + handle)
	if err != nil {
		log.Printf("error parsing url for %s with handle [%s]: %v", StoreName, handle, err)
		return "", false
	}
	cleanPageURL.RawQuery = url.Values{
		"utm_source": []string{config.UtmSource},
	}.Encode()
	return cleanPageURL.String(), true
}

func isSuggestProductFoil(title string, tags []string) bool {
	if strings.Contains(strings.ToLower(title), "[foil]") {
		return true
	}
	for _, tag := range tags {
		lower := strings.ToLower(strings.TrimSpace(tag))
		if lower == "foil" || strings.HasSuffix(lower, " foil") {
			return true
		}
	}
	return false
}
