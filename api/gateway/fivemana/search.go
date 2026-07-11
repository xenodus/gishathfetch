package fivemana

import (
	"context"
	"log"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const StoreName = "5 Mana"
const StoreBaseURL = "https://5-mana.sg"
const StoreSearchPath = "/search"

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

func (s Store) Search(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	var cards []gateway.Card

	// Build the request URL from constant components only;
	// user input is placed exclusively into query parameters via url.Values.
	apiURL := &url.URL{
		Scheme: "https",
		Host:   "5-mana.sg",
		Path:   StoreSearchPath,
		RawQuery: url.Values{
			"q":                     {searchStr},
			"filter.v.availability": {"1"},
		}.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return cards, err
	}
	if err := gateway.PrepareOutboundRequest(ctx, req, gateway.OutboundRequestOptions{
		Style:   gateway.OutboundStyleHTML,
		PageURL: apiURL,
	}); err != nil {
		return cards, err
	}
	client, err := gateway.NewOutboundHTTPClient(config.SearchAttemptTimeout)
	if err != nil {
		return cards, err
	}
	resp, err := client.Do(req)
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
		card, ok := parseProductCard(se, s.Name)
		if ok {
			cards = append(cards, card)
		}
	})

	return cards, nil
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
