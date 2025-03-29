package tefuda

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

const StoreName = "Tefuda"
const StoreBaseURL = "https://tefudagames.com"
const StoreSearchURL = "/search?q=%s"

// const binderposStoreURL = "bacc1b-3.myshopify.com"

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
	searchURL := s.BaseUrl + fmt.Sprintf(s.SearchUrl, url.QueryEscape(searchStr+" mtg"))
	var cards []gateway.Card

	c := colly.NewCollector()

	c.OnHTML("body", func(e *colly.HTMLElement) {
		e.ForEach("div.product-grid-container ul.product-grid li", func(_ int, el *colly.HTMLElement) {
			name := e.ChildText("div.product-card-wrapper > div > div.card__content > div.card__information > h3.card__heading a")
			link := e.ChildAttr("div.product-card-wrapper > div > div.card__content > div.card__information > h3.card__heading a", "href")
			img := e.ChildAttr("div.card__media img", "src")
			priceStr := e.ChildText("div.product-card-wrapper > div > div.card__content > div.card__information > div.card-information div.price__container > div.price__regular > span.price-item")
			price, err := parsePriceStr(priceStr)
			if err != nil {
				log.Printf("error parsing price for %s with value [%s]: %v", s.Name, priceStr, err)
				return
			}

			// url
			u := strings.TrimSpace(s.BaseUrl + link)
			cleanPageURL, err := url.Parse(u)
			if err != nil {
				log.Printf("error parsing url for %s with value [%s]: %v", s.Name, u, err)
				return
			}
			cleanPageURL.RawQuery = url.Values{
				"utm_source": []string{config.UtmSource},
			}.Encode()

			if price > 0 {
				cards = append(cards, gateway.Card{
					Name:    strings.TrimSpace(name),
					Url:     strings.TrimSpace(cleanPageURL.String()),
					InStock: true,
					Price:   price,
					Source:  s.Name,
					Img:     img,
				})
			}
		})
	})

	return cards, c.Visit(searchURL)
}

func parsePriceStr(priceStr string) (float64, error) {
	priceStr = strings.Replace(priceStr, "From", "", -1)
	priceStr = strings.Replace(priceStr, "$", "", -1)
	priceStr = strings.Replace(priceStr, "SGD", "", -1)
	priceStr = strings.TrimSpace(priceStr)

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price string: %w", err)
	}

	return price, nil
}
