package gog

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
	"mtg-price-checker-sg/pkg/config"
)

const StoreName = "Grey Ogre Games"
const StoreBaseURL = "https://www.greyogregames.com"
const StoreSearchURL = "/search?q="

const binderposStoreURL = "grey-ogre-games-singapore.myshopify.com"

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
		if httpStatusCode != http.StatusOK {
			log.Printf("falling back to scrap for [%s]", s.Name)
			return scrap(s, searchStr)
		}
		return cards, err
	}

	return cards, nil
}

type pagination struct {
	last int
	url  string
}

func scrap(s Store, searchStr string) ([]gateway.Card, error) {
	var (
		err   error
		cards []gateway.Card
	)

	pagination := new(pagination)
	searchURL := s.BaseUrl + s.SearchUrl + url.QueryEscape(searchStr)

	c := colly.NewCollector()

	c.OnHTML("body", func(e *colly.HTMLElement) {
		// get pagination
		e.ForEach("ol.pagination li", func(_ int, el *colly.HTMLElement) {
			elStr := strings.Replace(el.Text, "«", "", -1)
			elStr = strings.Replace(elStr, "page", "", -1)
			elStr = strings.Replace(elStr, "Next", "", -1)
			elStr = strings.Replace(elStr, "Previous", "", -1)
			elStr = strings.Replace(elStr, "»", "", -1)
			elStr = strings.TrimSpace(elStr)
			if elStr != "" && elStr != "1" && el.ChildAttr("a", "href") != "" {
				elInt, strConvErr := strconv.Atoi(elStr)
				if strConvErr == nil {
					pagination.last = elInt
					pagination.url = el.ChildAttr("a", "href")
				}
			}
		})

		// get cards
		e.ForEach("div.productCard__card", func(_ int, el *colly.HTMLElement) {
			var (
				isInstock bool
				price     float64
			)

			// in stock
			if len(el.ChildTexts("div.productCard__button--outOfStock")) == 0 {
				isInstock = true
			}

			if isInstock {
				el.ForEach("ul.productChip__grid li", func(_ int, el2 *colly.HTMLElement) {
					if el2.Attr("data-variantavailable") == "true" && el2.Attr("data-variantqty") != "0" {
						priceStr := el2.Attr("data-variantprice")
						priceStr = strings.Replace(priceStr, "$", "", -1)
						priceStr = strings.Replace(priceStr, ",", "", -1)
						priceStr = strings.Replace(priceStr, "SGD", "", -1)
						price, _ = strconv.ParseFloat(strings.TrimSpace(priceStr), 64)
						price = price / 100

						cardUrl := strings.TrimSpace(s.BaseUrl + el.ChildAttr("a", "href"))
						u, err := url.Parse(strings.TrimSpace(s.BaseUrl + cardUrl))
						if err != nil {
							log.Printf("error parsing url for %s with value [%s]: %v", s.Name, cardUrl, err)
							return
						}
						cleanPageURL := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)

						if price > 0 {
							cards = append(cards, gateway.Card{
								Name:       strings.TrimSpace(el.ChildText("p.productCard__title")),
								Url:        strings.TrimSpace(cleanPageURL),
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
		})
	})

	err = c.Visit(searchURL)
	if err != nil {
		return []gateway.Card{}, err
	}

	if pagination.url != "" {
		log.Println("Pagination exists for " + s.Name)

		c2 := colly.NewCollector()

		for i := 2; i <= pagination.last; i++ {
			searchURL = s.BaseUrl + strings.Replace(pagination.url, "page="+strconv.Itoa(pagination.last), "page="+strconv.Itoa(i), 1)

			c2.OnHTML("div.collectionGrid", func(e *colly.HTMLElement) {
				e.ForEach("div.productCard__card", func(_ int, el *colly.HTMLElement) {
					var (
						isInstock bool
						price     float64
					)

					// in stock
					if len(el.ChildTexts("div.productCard__button--outOfStock")) == 0 {
						isInstock = true
					}

					if isInstock {
						el.ForEach("ul.productChip__grid li", func(_ int, el2 *colly.HTMLElement) {
							if el2.Attr("data-variantavailable") == "true" && el2.Attr("data-variantqty") != "0" {
								priceStr := el2.Attr("data-variantprice")
								priceStr = strings.Replace(priceStr, "$", "", -1)
								priceStr = strings.Replace(priceStr, ",", "", -1)
								priceStr = strings.Replace(priceStr, "SGD", "", -1)
								price, _ = strconv.ParseFloat(strings.TrimSpace(priceStr), 64)
								price = price / 100

								cardUrl := strings.TrimSpace(s.BaseUrl + el.ChildAttr("a", "href"))
								u, err := url.Parse(strings.TrimSpace(s.BaseUrl + cardUrl))
								if err != nil {
									log.Printf("error parsing url for %s with value [%s]: %v", s.Name, cardUrl, err)
									return
								}
								cleanPageURL := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)

								if price > 0 {
									cards = append(cards, gateway.Card{
										Name:       strings.TrimSpace(el.ChildText("p.productCard__title")),
										Url:        strings.TrimSpace(cleanPageURL),
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
				})
			})

			log.Println("Searching page no: ", i)
			log.Println(searchURL)

			err = c2.Visit(searchURL)
			if err != nil {
				break
			}

			// Application's max page limit
			if i >= config.MaxPagesToSearch {
				break
			}
		}
	}

	return cards, err
}
