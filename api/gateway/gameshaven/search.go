package gameshaven

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/config"
)

const StoreName = "Games Haven"
const StoreBaseURL = "https://www.gameshaventcg.com"
const StoreSearchURL = "/search?q="

const binderposStoreURL = "games-haven-sg.myshopify.com"

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
	return scrap(s, searchStr)
}

func scrap(s Store, searchStr string) ([]gateway.Card, error) {
	searchURL := s.BaseUrl + s.SearchUrl + url.QueryEscape(searchStr+" mtg")
	var cards []gateway.Card

	c := colly.NewCollector()

	c.OnHTML("div.collectionGrid", func(e *colly.HTMLElement) {
		e.ForEach("div.productCard__card", func(_ int, el *colly.HTMLElement) {
			var (
				isInstock bool
				price     float64
			)

			// not out of stock
			if el.ChildText("form") != "" {
				isInstock = true

				if isInstock {
					el.ForEach("ul.productChip__grid li", func(_ int, el2 *colly.HTMLElement) {
						if el2.Attr("data-variantavailable") == "true" && el2.Attr("data-variantqty") != "0" {
							priceStr := el2.Attr("data-variantprice")
							priceStr = strings.Replace(priceStr, "$", "", -1)
							priceStr = strings.Replace(priceStr, ",", "", -1)
							priceStr = strings.Replace(priceStr, "SGD", "", -1)
							price, _ = strconv.ParseFloat(strings.TrimSpace(priceStr), 64)
							price = price / 100

							// url with variant (quality)
							u := strings.TrimSpace(el.ChildAttr("a", "href"))
							cleanPageURL, err := url.Parse(u)
							if err != nil {
								log.Printf("error parsing url for %s with value [%s]: %v", s.Name, u, err)
								return
							}
							cleanPageURL.RawQuery = url.Values{
								"variant":    []string{el2.Attr("data-variantid")},
								"utm_source": []string{config.UtmSource},
							}.Encode()

							if price > 0 {
								cards = append(cards, gateway.Card{
									Name:      strings.TrimSpace(el.ChildText("p.productCard__title")),
									Url:       strings.TrimSpace(cleanPageURL.String()),
									InStock:   isInstock,
									Price:     price,
									Source:    s.Name,
									Img:       strings.TrimSpace("https:" + el.ChildAttr("img", "data-src")),
									Quality:   el2.Attr("data-varianttitle"),
									ExtraInfo: []string{fmt.Sprintf("[%s]", el.ChildText("p.productCard__setName"))},
								})
							}
						}
					})
				}
			}
		})
	})

	return cards, c.Visit(searchURL)
}
