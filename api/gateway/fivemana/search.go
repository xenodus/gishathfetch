package fivemana

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"

	"github.com/PuerkitoBio/goquery"
)

const StoreName = "5 Mana"
const StoreBaseURL = "https://5-mana.sg"
const StoreSearchPath = "/search"

// storefrontSearchSectionID is Dawn's main search results section. Requesting it
// alone returns the product grid without the full theme chrome.
const storefrontSearchSectionID = "main-search"

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
	cards, err := s.searchGraphQL(ctx, searchStr)
	if err == nil && gateway.CardsMatchSearch(cards, searchStr) {
		return cards, nil
	}
	if err == nil && len(cards) > 0 {
		log.Printf("%s: graphql results do not match %q, falling back to HTML", s.Name, searchStr)
	} else if err != nil {
		if gateway.IsHTTPServerError(err) {
			return nil, err
		}
		log.Printf("%s: graphql search failed, falling back to HTML: %v", s.Name, err)
	}

	htmlCards, htmlErr := s.searchHTML(ctx, searchStr)
	if htmlErr != nil {
		return nil, fmt.Errorf("graphql: %w; html: %v", err, htmlErr)
	}
	return htmlCards, nil
}

func (s Store) searchHTML(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	var cards []gateway.Card

	storeBase, err := s.storeBaseURL()
	if err != nil {
		return cards, err
	}

	apiURL := &url.URL{
		Scheme: storeBase.Scheme,
		Host:   storeBase.Host,
		Path:   StoreSearchPath,
		RawQuery: url.Values{
			"q":                     {searchStr},
			"filter.v.availability": {"1"},
			"section_id":            {storefrontSearchSectionID},
		}.Encode(),
	}

	resp, err := gateway.DoOutboundGET(ctx, apiURL.String(), fiveManaOutboundOpts(storeBase, apiURL, gateway.OutboundStyleHTML), config.SearchAttemptTimeout)
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

	doc.Find("ul.product-grid li").Each(func(_ int, se *goquery.Selection) {
		card, ok := parseProductCard(se, s.Name)
		if ok {
			cards = append(cards, card)
		}
	})

	return cards, nil
}

func fiveManaOutboundOpts(storeBase *url.URL, pageURL *url.URL, style gateway.OutboundRequestStyle) gateway.OutboundRequestOptions {
	opts := gateway.OutboundRequestOptions{
		Style:                  style,
		PageURL:                pageURL,
		StoreBase:              storeBase,
		ShopifySGDCurrency: true,
		SkipDirect:         true,
	}
	// Non-production hosts (httptest unit tests) must use the direct transport.
	if storeBase == nil || storeBase.Host != "5-mana.sg" {
		opts.SkipDirect = false
	}
	return opts
}

func (s Store) storeBaseURL() (*url.URL, error) {
	return url.Parse(s.BaseUrl)
}

func parseProductCard(se *goquery.Selection, storeName string) (gateway.Card, bool) {
	// Shopify's availability filter can still surface sold-out listings; the theme
	// marks them with price--sold-out and a "Sold out" badge.
	if se.Find("div.price.price--sold-out").Length() > 0 {
		return gateway.Card{}, false
	}

	heading := se.Find("h3.card__heading.h5 a")
	rawName := strings.TrimSpace(heading.Text())
	name, isFoil := parseNameAndFoil(rawName)
	if name == "" {
		return gateway.Card{}, false
	}

	price, err := util.ParsePrice(se.Find("span.price-item.price-item--sale.price-item--last").Text())
	if err != nil || price <= 0 {
		return gateway.Card{}, false
	}

	cardURL, err := productURLWithUTM(StoreBaseURL + se.Find("h3.card__heading a").AttrOr("href", ""))
	if err != nil {
		log.Printf("error parsing url for %s with value [%s]: %v", storeName, heading.AttrOr("href", ""), err)
		return gateway.Card{}, false
	}

	return gateway.Card{
		Name:      name,
		Url:       cardURL,
		Img:       se.Find("div.card__media img").AttrOr("src", ""),
		InStock:   true,
		IsFoil:    isFoil,
		Price:     price,
		Source:    storeName,
		ExtraInfo: extraInfoFromTitle(rawName),
	}, true
}

func productURLWithUTM(raw string) (string, error) {
	cleanPageURL, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	if cleanPageURL.Host == "" {
		cleanPageURL, err = url.Parse(StoreBaseURL + cleanPageURL.Path)
		if err != nil {
			return "", err
		}
	}
	q := cleanPageURL.Query()
	q.Set("utm_source", config.UtmSource)
	cleanPageURL.RawQuery = q.Encode()
	return cleanPageURL.String(), nil
}
