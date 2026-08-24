package tefuda

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"

	"github.com/PuerkitoBio/goquery"
)

const StoreName = "Tefuda"
const StoreBaseURL = "https://tefudagames.com"
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
	cards, err := s.searchGraphQL(ctx, searchStr)
	if err == nil {
		return cards, nil
	}
	if gateway.IsHTTPServerError(err) {
		return nil, err
	}

	slog.Warn("graphql search failed, falling back to HTML", "store", s.Name, "err", err)

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
			"q":    {mtgSinglesSearchQuery(searchStr)},
			"type": {"product"},
		}.Encode(),
	}

	resp, err := gateway.DoOutboundGET(ctx, apiURL.String(), tefudaOutboundOpts(storeBase, apiURL, gateway.OutboundStyleHTML), config.SearchAttemptTimeout)
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

func tefudaOutboundOpts(storeBase *url.URL, pageURL *url.URL, style gateway.OutboundRequestStyle) gateway.OutboundRequestOptions {
	return gateway.OutboundRequestOptions{
		Style:              style,
		PageURL:            pageURL,
		StoreBase:          storeBase,
		ShopifySGDCurrency: true,
	}
}

func (s Store) storeBaseURL() (*url.URL, error) {
	return url.Parse(s.BaseUrl)
}

func mtgSinglesSearchQuery(searchStr string) string {
	return fmt.Sprintf(`product_type:"%s" AND %s`, storefrontMTGType, strings.TrimSpace(searchStr))
}

func parseProductCard(se *goquery.Selection, storeName string) (gateway.Card, bool) {
	if se.Find("div.price.price--sold-out").Length() > 0 {
		return gateway.Card{}, false
	}

	heading := se.Find("h3.card__heading a").First()
	rawName := strings.TrimSpace(heading.Text())
	name, isFoil := parseNameAndFoil(rawName)
	if name == "" {
		return gateway.Card{}, false
	}

	priceText := strings.TrimSpace(se.Find("span.price-item.price-item--sale.price-item--last").Text())
	if priceText == "" {
		priceText = strings.TrimSpace(se.Find("span.price-item.price-item--regular").First().Text())
	}
	price, err := util.ParsePrice(priceText)
	if err != nil || price <= 0 {
		return gateway.Card{}, false
	}

	cardURL, err := productURLWithUTM(StoreBaseURL + heading.AttrOr("href", ""))
	if err != nil {
		slog.Warn("error parsing url", "store", storeName, "value", heading.AttrOr("href", ""), "err", err)
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
