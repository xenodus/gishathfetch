package fivemana

import (
	"fmt"
	"log"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/config"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const StoreName = "5 Mana"
const StoreBaseURL = "https://5-mana.sg"
const StoreSearchURL = "/search?q=%s&filter.v.availability=1"

type Store struct {
	Name      string
	BaseUrl   string
	SearchUrl string
}

func NewLGS() gateway.LGS {
	return Store{
		Name:      StoreName,
		BaseUrl:   StoreBaseURL,
		SearchUrl: StoreSearchURL,
	}
}

func (s Store) Search(searchStr string) ([]gateway.Card, error) {
	var cards []gateway.Card
	apiURL := s.BaseUrl + fmt.Sprintf(s.SearchUrl, url.QueryEscape(searchStr))

	resp, err := http.Get(apiURL)
	if err != nil {
		return cards, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return cards, err
	}

	doc.Find("ul.product-grid li").Each(func(i int, se *goquery.Selection) {
		c := gateway.Card{
			Source: s.Name,
		}

		name := se.Find("h3.card__heading.h5 a").Text()

		squareBracketIndex := strings.Index(name, "[")
		if squareBracketIndex >= 0 {
			c.Name = strings.TrimSpace(name[:squareBracketIndex-1])
			c.ExtraInfo = append(c.ExtraInfo, strings.TrimSpace(name[squareBracketIndex:]))
		}

		c.Url = StoreBaseURL + se.Find("h3.card__heading a").AttrOr("href", "")
		c.Img = se.Find("div.card__media img").AttrOr("src", "")
		c.InStock = true

		price, err := parsePrice(se.Find("span.price-item.price-item--sale.price-item--last").Text())
		if err != nil {
			c.InStock = false
		}
		c.Price = price

		// url
		cleanPageURL, err := url.Parse(c.Url)
		if err != nil {
			log.Printf("error parsing url for %s with value [%s]: %v", s.Name, c.Url, err)
			return
		}
		cleanPageURL.RawQuery = url.Values{
			"utm_source": []string{config.UtmSource},
		}.Encode()
		c.Url = cleanPageURL.String()

		if c.Name != "" && c.InStock {
			cards = append(cards, c)
		}
	})

	return cards, nil
}

func parsePrice(price string) (float64, error) {
	priceStr := strings.Replace(price, "$", "", -1)
	priceStr = strings.Replace(priceStr, "SGD", "", -1)
	priceStr = strings.TrimSpace(priceStr)
	return strconv.ParseFloat(priceStr, 64)
}
