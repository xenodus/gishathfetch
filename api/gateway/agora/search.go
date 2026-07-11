package agora

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"

	"github.com/PuerkitoBio/goquery"
)

const StoreName = "Agora Hobby"
const StoreBaseURL = "https://agorahobby.com"
const StoreSearchPath = "/store/search"
const storeCategoryMTG = "mtg"

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

	baseURL, err := url.Parse(s.BaseUrl)
	if err != nil {
		return cards, err
	}

	apiURL := &url.URL{
		Scheme: baseURL.Scheme,
		Host:   baseURL.Host,
		Path:   s.SearchPath,
		RawQuery: url.Values{
			"category":    {storeCategoryMTG},
			"searchfield": {searchStr},
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

	client, err := gateway.NewOutboundHTTPClient(config.AgoraSearchAttemptTimeout)
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

	doc.Find("div#store_listingcontainer div.store-item").Each(func(_ int, se *goquery.Selection) {
		card, ok := parseStoreItem(se, s.Name, apiURL, searchStr)
		if ok {
			cards = append(cards, card)
		}
	})

	return cards, nil
}

func parseStoreItem(se *goquery.Selection, storeName string, apiURL *url.URL, searchStr string) (gateway.Card, bool) {
	if se.Find("div.store-item-stock").Text() == "Stock: 0" {
		return gateway.Card{}, false
	}

	priceStr := strings.TrimSpace(se.Find("div.store-item-price").Text())
	priceStr = strings.Replace(priceStr, "$", "", -1)
	priceStr = strings.Replace(priceStr, ",", "", -1)
	price, err := strconv.ParseFloat(strings.TrimSpace(priceStr), 64)
	if err != nil || price <= 0 {
		return gateway.Card{}, false
	}

	categoryText := se.Find("div.store-item-cat").Text()
	quality := ""
	qualityParts := strings.Split(categoryText, " - ")
	if len(qualityParts) == 2 {
		quality = strings.TrimSpace(qualityParts[1])
	}

	var extraInfo []string
	if strings.Contains(categoryText, "]") {
		set := categoryText[:strings.Index(categoryText, "]")+1]
		extraInfo = append(extraInfo, set)
	}

	name := strings.TrimSpace(se.Find("div.store-item-title").Text())
	if name == "" {
		return gateway.Card{}, false
	}

	cardURL, err := buildCardURL(apiURL, searchStr)
	if err != nil {
		log.Printf("error parsing url for %s with value [%s]: %v", storeName, apiURL.String(), err)
		return gateway.Card{}, false
	}

	return gateway.Card{
		Name:      name,
		Url:       cardURL,
		InStock:   true,
		IsFoil:    strings.Contains(name, "FOIL"),
		Price:     price,
		Source:    storeName,
		Img:       strings.TrimSpace(se.Find("div.store-item-img").AttrOr("data-img", "")),
		Quality:   util.MapQuality(quality),
		ExtraInfo: extraInfo,
	}, true
}

func buildCardURL(apiURL *url.URL, searchStr string) (string, error) {
	cleanPageURL, err := url.Parse(apiURL.String())
	if err != nil {
		return "", err
	}
	cleanPageURL.RawQuery = url.Values{
		"category":    []string{storeCategoryMTG},
		"searchfield": []string{searchStr},
		"utm_source":  []string{config.UtmSource},
	}.Encode()
	return cleanPageURL.String(), nil
}
