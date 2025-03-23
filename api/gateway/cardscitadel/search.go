package cardscitadel

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/binderpos"
)

const StoreName = "Cards Citadel"
const StoreBaseURL = "https://cardscitadel.com/"
const StoreSearchURL = "/search?q=*%s*"

const binderposStoreURL = "card-citadel.myshopify.com"

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
	reqPayload, err := json.Marshal(binderpos.Payload{
		StoreURL:    binderposStoreURL,
		Game:        binderpos.ProductTypeMTG.ToString(),
		Title:       searchStr,
		InstockOnly: true,
	})
	if err != nil {
		return []gateway.Card{}, err
	}

	cards, httpStatusCode, err := binderpos.GetCards(s.Name, s.BaseUrl, reqPayload)
	if err != nil {
		return cards, err
	}

	if httpStatusCode != http.StatusOK {
		log.Printf("falling back to scrap for [%s]", s.Name)
		return scrap(s, searchStr)
	}

	return cards, nil
}

func scrap(s Store, searchStr string) ([]gateway.Card, error) {
	searchURL := s.BaseUrl + fmt.Sprintf(s.SearchUrl, url.QueryEscape(searchStr))
	var cards []gateway.Card

	c := colly.NewCollector()

	c.OnHTML("div.container", func(e *colly.HTMLElement) {
		e.ForEach("div.Norm", func(_ int, el *colly.HTMLElement) {
			var isInstock bool

			if len(el.ChildTexts("div.addNow")) > 0 {
				for i := 0; i < len(el.ChildTexts("div.addNow")); i++ {
					isInstock = el.ChildTexts("div.addNow")[i] != ""

					if isInstock {
						priceStr := strings.TrimSpace(el.ChildTexts("div.addNow")[i])

						price, quality, err := parsePriceAndQuality(priceStr)
						if err != nil {
							continue
						}

						if price > 0 {
							cards = append(cards, gateway.Card{
								Name:    strings.TrimSpace(el.ChildText("p.productTitle")),
								Url:     strings.TrimSpace(s.BaseUrl + strings.Replace(el.ChildAttr("a", "href"), "/products/", "products/", -1)),
								InStock: isInstock,
								Price:   price,
								Source:  s.Name,
								Img:     strings.TrimSpace("https:" + el.ChildAttr("img", "src")),
								Quality: quality,
							})
						}
					}
				}
			}
		})
	})

	return cards, c.Visit(searchURL)
}

func parsePriceAndQuality(priceQualityStr string) (float64, string, error) {
	priceQualityStrSlice := strings.Split(priceQualityStr, " - ")
	if len(priceQualityStrSlice) == 2 {
		quality := strings.TrimSpace(priceQualityStrSlice[0])
		price, err := parsePrice(priceQualityStrSlice[1])
		return price, quality, err
	}
	return 0, "", nil
}

func parsePrice(price string) (float64, error) {
	priceStr := strings.TrimSpace(price)
	priceStr = strings.Replace(priceStr, "$", "", -1)
	priceStr = strings.Replace(priceStr, ",", "", -1)
	return strconv.ParseFloat(priceStr, 64)
}
