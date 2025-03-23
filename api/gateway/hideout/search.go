package hideout

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"mtg-price-checker-sg/gateway"
)

const StoreName = "Hideout"
const StoreBaseURL = "https://hideoutcg.com"
const StoreSearchURL = "/search?q="

const binderposStoreURL = "220022-20.myshopify.com"

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

//func (s Store) Search(searchStr string) ([]gateway.Card, error) {
//	reqPayload, err := json.Marshal(binderpos.Payload{
//		StoreURL:    binderposStoreURL,
//		Game:        binderpos.ProductTypeMTG.ToString(),
//		Title:       searchStr,
//		InstockOnly: true,
//	})
//	if err != nil {
//		return []gateway.Card{}, err
//	}
//
//	return binderpos.GetCards(s.Name, s.BaseUrl, reqPayload)
//}

func (s Store) Search(searchStr string) ([]gateway.Card, error) {
	searchURL := s.BaseUrl + s.SearchUrl + url.QueryEscape(searchStr)
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

							if price > 0 {
								cards = append(cards, gateway.Card{
									Name:       strings.TrimSpace(el.ChildText("p.productCard__title")),
									Url:        strings.TrimSpace(s.BaseUrl + el.ChildAttr("a", "href")),
									InStock:    isInstock,
									Price:      price,
									Source:     s.Name,
									Img:        strings.TrimSpace("https:" + el.ChildAttr("img", "data-src")),
									Quality:    el2.Attr("data-varianttitle"),
									IsScrapped: true,
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
