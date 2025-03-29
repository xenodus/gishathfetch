package agora

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

const StoreName = "Agora Hobby"
const StoreBaseURL = "https://agorahobby.com"
const StoreSearchURL = "/store/search?category=mtg&searchfield=%s"

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
	searchURL := s.BaseUrl + fmt.Sprintf(s.SearchUrl, url.QueryEscape(searchStr))
	var cards []gateway.Card

	c := colly.NewCollector()

	c.OnHTML("div#store_listingcontainer", func(e *colly.HTMLElement) {
		e.ForEach("div.store-item", func(_ int, el *colly.HTMLElement) {
			var (
				isInstock bool
				price     float64
				quality   string
			)

			// in stock
			if el.ChildText("div.store-item-stock") != "Stock: 0" {
				isInstock = true
			}

			// price
			priceStr := strings.TrimSpace(el.ChildText("div.store-item-price"))
			priceStr = strings.Replace(priceStr, "$", "", -1)
			priceStr = strings.Replace(priceStr, ",", "", -1)
			price, _ = strconv.ParseFloat(strings.TrimSpace(priceStr), 64)

			// quality
			qualityStr := el.ChildText("div.store-item-cat")
			qualityStrSlice := strings.Split(qualityStr, " - ")
			if len(qualityStrSlice) == 2 {
				quality = strings.TrimSpace(qualityStrSlice[1])
			}

			var (
				extraInfo []string
				set       string
			)
			if strings.Index(el.ChildText("div.store-item-cat"), "]") != -1 {
				set = el.ChildText("div.store-item-cat")[:strings.Index(el.ChildText("div.store-item-cat"), "]")+1]
				extraInfo = append(extraInfo, set)
			}

			// name
			name := el.ChildText("div.store-item-title")

			// url
			cleanPageURL, err := url.Parse(strings.TrimSpace(searchURL))
			if err != nil {
				log.Printf("error parsing url for %s with value [%s]: %v", s.Name, searchURL, err)
				return
			}
			cleanPageURL.RawQuery = url.Values{
				"category":    []string{"mtg"},
				"searchfield": []string{url.QueryEscape(searchStr)},
				"utm_source":  []string{config.UtmSource},
			}.Encode()

			// Exclude Japanese cards
			if price > 0 && !strings.Contains(name, "Japanese") {
				cards = append(cards, gateway.Card{
					Name:      strings.TrimSpace(el.ChildText("div.store-item-title")),
					Url:       strings.TrimSpace(cleanPageURL.String()),
					InStock:   isInstock,
					Price:     price,
					Source:    s.Name,
					Img:       strings.TrimSpace(el.ChildAttr("div.store-item-img", "data-img")),
					Quality:   quality,
					ExtraInfo: extraInfo,
				})
			}
		})
	})

	return cards, c.Visit(searchURL)
}
