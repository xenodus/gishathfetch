package fivemana

import (
	"context"
	"encoding/json"
	"log"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/config"
	"net/url"
	"strconv"
	"strings"
)

const StoreName = "5 Mana"
const StoreBaseURL = "https://5-mana.sg"
const StoreSearchPath = "/search/suggest.json"
const suggestProductLimit = "20"

type Store struct {
	Name       string
	BaseUrl    string
	SearchPath string
}

func NewLGS() gateway.LGS {
	return Store{
		Name:       StoreName,
		BaseUrl:    StoreBaseURL,
		SearchPath: StoreSearchPath,
	}
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
	Title     string   `json:"title"`
	Price     string   `json:"price"`
	Image     string   `json:"image"`
	URL       string   `json:"url"`
	Tags      []string `json:"tags"`
}

func (s Store) Search(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	var cards []gateway.Card

	storeBase, err := url.Parse(s.BaseUrl)
	if err != nil {
		return cards, err
	}

	apiURL := &url.URL{
		Scheme: "https",
		Host:   "5-mana.sg",
		Path:   StoreSearchPath,
		RawQuery: url.Values{
			"q":                  {searchStr},
			"resources[type]":    {"product"},
			"resources[limit]": {suggestProductLimit},
		}.Encode(),
	}

	resp, err := gateway.DoOutboundGET(ctx, apiURL.String(), gateway.OutboundRequestOptions{
		Style:              gateway.OutboundStyleJSON,
		PageURL:            storeBase,
		ShopifySGDCurrency: true,
		// Shopify storefronts behind Cloudflare respond better to browser-like
		// requests than signed bot traffic on the suggest API path.
		SkipWebBotAuth: true,
	}, config.SearchAttemptTimeout)
	if err != nil {
		return cards, err
	}
	defer resp.Body.Close()

	body, err := gateway.ReadResponseBody(resp)
	if err != nil {
		return cards, err
	}

	var payload suggestResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return cards, err
	}

	for _, product := range payload.Resources.Results.Products {
		card, ok := parseSuggestProduct(product, s.Name)
		if ok {
			cards = append(cards, card)
		}
	}

	return cards, nil
}

func parseSuggestProduct(product suggestProduct, storeName string) (gateway.Card, bool) {
	if !product.Available {
		return gateway.Card{}, false
	}

	title := strings.TrimSpace(product.Title)
	if title == "" {
		return gateway.Card{}, false
	}

	price, err := strconv.ParseFloat(strings.TrimSpace(product.Price), 64)
	if err != nil || price <= 0 {
		return gateway.Card{}, false
	}

	productURL := strings.TrimSpace(product.URL)
	if productURL == "" {
		return gateway.Card{}, false
	}
	if strings.HasPrefix(productURL, "/") {
		productURL = StoreBaseURL + productURL
	}

	cleanPageURL, err := url.Parse(productURL)
	if err != nil {
		log.Printf("error parsing url for %s with value [%s]: %v", storeName, productURL, err)
		return gateway.Card{}, false
	}
	cleanPageURL.RawQuery = url.Values{
		"utm_source": []string{config.UtmSource},
	}.Encode()

	return gateway.Card{
		Name:      strings.TrimSpace(strings.Replace(title, "[Foil]", "", -1)),
		Url:       cleanPageURL.String(),
		Img:       strings.TrimSpace(product.Image),
		InStock:   true,
		IsFoil:    isFoilProduct(title, product.Tags),
		Price:     price,
		Source:    storeName,
	}, true
}

func isFoilProduct(title string, tags []string) bool {
	if strings.Contains(strings.ToLower(title), "[foil]") {
		return true
	}
	for _, tag := range tags {
		if strings.EqualFold(strings.TrimSpace(tag), "foil") {
			return true
		}
	}
	return false
}
