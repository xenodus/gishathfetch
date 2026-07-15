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

	"github.com/PuerkitoBio/goquery"
)

const StoreName = "5 Mana"
const StoreBaseURL = "https://5-mana.sg"
const StoreSearchPath = "/search"
const StoreSuggestPath = "/search/suggest.json"
const suggestProductLimit = "10"

type Store struct {
	Name        string
	BaseUrl     string
	SearchPath  string
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
		SearchPath:  StoreSearchPath,
		SuggestPath: StoreSuggestPath,
	}
}

func (s Store) Search(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	cards, err := searchBySuggestJSON(ctx, s.Name, searchStr)
	if err == nil && len(cards) > 0 {
		return cards, nil
	}
	return searchByHTML(ctx, s.Name, searchStr)
}

func searchBySuggestJSON(ctx context.Context, storeName, searchStr string) ([]gateway.Card, error) {
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
			"resources[type]":                          {"product"},
			"resources[limit]":                         {suggestProductLimit},
			"resources[options][unavailable_products]": {"hide"},
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

	return mapSuggestProductsToCards(parsed.Resources.Results.Products, storeName), nil
}

func searchByHTML(ctx context.Context, storeName, searchStr string) ([]gateway.Card, error) {
	var cards []gateway.Card

	apiURL := &url.URL{
		Scheme: "https",
		Host:   "5-mana.sg",
		Path:   StoreSearchPath,
		RawQuery: url.Values{
			"q":                     {searchStr},
			"filter.v.availability": {"1"},
		}.Encode(),
	}

	resp, err := gateway.DoOutboundGET(ctx, apiURL.String(), gateway.OutboundRequestOptions{
		Style:              gateway.OutboundStyleHTML,
		PageURL:            apiURL,
		ShopifySGDCurrency: true,
		SkipDirect:         true,
	}, config.SearchAttemptTimeout)
	if err != nil {
		return cards, err
	}
	defer resp.Body.Close()

	body, err := gateway.ReadResponseBody(resp)
	if err != nil {
		return cards, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return cards, err
	}

	doc.Find("ul.product-grid li").Each(func(i int, se *goquery.Selection) {
		card, ok := parseProductCard(se, storeName)
		if ok {
			cards = append(cards, card)
		}
	})

	return cards, nil
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

func parseProductCard(se *goquery.Selection, storeName string) (gateway.Card, bool) {
	// Shopify's availability filter can still surface sold-out listings; the theme
	// marks them with price--sold-out and a "Sold out" badge.
	if se.Find("div.price.price--sold-out").Length() > 0 {
		return gateway.Card{}, false
	}

	c := gateway.Card{
		Source: storeName,
	}

	// name e.g. Rhystic Study (Anime Borderless) [Wilds of Eldraine: Enchanting Tales] [Foil]
	heading := se.Find("h3.card__heading.h5 a")
	c.Name = strings.TrimSpace(strings.Replace(heading.Text(), "[Foil]", "", -1))
	c.IsFoil = strings.Contains(strings.ToLower(heading.Text()), "[foil]")

	c.Url = StoreBaseURL + se.Find("h3.card__heading a").AttrOr("href", "")
	c.Img = se.Find("div.card__media img").AttrOr("src", "")
	c.InStock = true

	price, err := util.ParsePrice(se.Find("span.price-item.price-item--sale.price-item--last").Text())
	if err != nil {
		c.InStock = false
	}
	c.Price = price

	cleanPageURL, err := url.Parse(c.Url)
	if err != nil {
		log.Printf("error parsing url for %s with value [%s]: %v", storeName, c.Url, err)
		return gateway.Card{}, false
	}
	cleanPageURL.RawQuery = url.Values{
		"utm_source": []string{config.UtmSource},
	}.Encode()
	c.Url = cleanPageURL.String()

	return c, c.Name != "" && c.InStock
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
