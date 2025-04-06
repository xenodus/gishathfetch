package binderpos

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/config"
)

func (i impl) Scrap(scrapVariant int, storeName, baseUrl, searchUrl, searchStr string) ([]gateway.Card, error) {
	switch scrapVariant {
	case 1:
		return scrapVariant1(storeName, baseUrl, searchUrl, searchStr)
	case 2:
		return scrapVariant2(storeName, baseUrl, searchUrl, searchStr)
	case 3:
		return scrapVariant3(storeName, baseUrl, searchUrl, searchStr)
	case 4:
		return scrapVariant4(storeName, baseUrl, searchUrl, searchStr)
	}
	return []gateway.Card{}, fmt.Errorf("invalid scrap variant: %d", scrapVariant)
}

// tefuda
func scrapVariant4(storeName, baseUrl, searchUrl, searchStr string) ([]gateway.Card, error) {
	searchURL := baseUrl + fmt.Sprintf(searchUrl, url.QueryEscape(searchStr+" mtg"))
	var cards []gateway.Card

	c := colly.NewCollector()

	c.OnHTML("body", func(e *colly.HTMLElement) {
		e.ForEach("div.product-grid-container ul.product-grid li", func(_ int, el *colly.HTMLElement) {
			name := e.ChildText("div.product-card-wrapper > div > div.card__content > div.card__information > h3.card__heading a")
			link := e.ChildAttr("div.product-card-wrapper > div > div.card__content > div.card__information > h3.card__heading a", "href")
			img := e.ChildAttr("div.card__media img", "src")
			priceStr := e.ChildText("div.product-card-wrapper > div > div.card__content > div.card__information > div.card-information div.price__container > div.price__regular > span.price-item")
			price, err := parsePrice(priceStr)
			if err != nil {
				log.Printf("error parsing price for %s with value [%s]: %v", storeName, priceStr, err)
				return
			}

			// url
			u := strings.TrimSpace(baseUrl + link)
			cleanPageURL, err := url.Parse(u)
			if err != nil {
				log.Printf("error parsing url for %s with value [%s]: %v", storeName, u, err)
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
					Source:  storeName,
					Img:     img,
				})
			}
		})
	})

	return cards, c.Visit(searchURL)
}

type pagination struct {
	last int
	url  string
}

// scrap with pagination
// games haven
// gog
// hideout
func scrapVariant3(storeName, baseUrl, searchUrl, searchStr string) ([]gateway.Card, error) {
	var (
		err   error
		cards []gateway.Card
	)

	page := new(pagination)
	searchURL := baseUrl + fmt.Sprintf(searchUrl, url.QueryEscape(searchStr+" mtg"))

	c := colly.NewCollector()

	c.OnHTML("body", func(e *colly.HTMLElement) {
		// get page
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
					page.last = elInt
					page.url = el.ChildAttr("a", "href")
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

						u := strings.TrimSpace(baseUrl + el.ChildAttr("a", "href"))
						cleanPageURL, err := url.Parse(u)
						if err != nil {
							log.Printf("error parsing url for %s with value [%s]: %v", storeName, u, err)
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
								Source:    storeName,
								Img:       strings.TrimSpace("https:" + el.ChildAttr("img", "data-src")),
								Quality:   el2.Attr("data-varianttitle"),
								ExtraInfo: []string{el.ChildText("p.productCard__setName")},
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

	if page.url != "" {
		log.Println("Pagination exists for " + storeName)

		c2 := colly.NewCollector()

		for i := 2; i <= page.last; i++ {
			searchURL = baseUrl + strings.Replace(page.url, "page="+strconv.Itoa(page.last), "page="+strconv.Itoa(i), 1)

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

								u := strings.TrimSpace(baseUrl + el.ChildAttr("a", "href"))
								cleanPageURL, err := url.Parse(u)
								if err != nil {
									log.Printf("error parsing url for %s with value [%s]: %v", storeName, u, err)
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
										Source:    storeName,
										Img:       strings.TrimSpace("https:" + el.ChildAttr("img", "data-src")),
										Quality:   el2.Attr("data-varianttitle"),
										ExtraInfo: []string{el.ChildText("p.productCard__setName")},
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

	//c.OnHTML("div.collectionGrid", func(e *colly.HTMLElement) {
	//	e.ForEach("div.productCard__card", func(_ int, el *colly.HTMLElement) {
	//		var (
	//			isInstock bool
	//			price     float64
	//		)
	//
	//		// not out of stock
	//		if el.ChildText("form") != "" {
	//			isInstock = true
	//
	//			if isInstock {
	//				el.ForEach("ul.productChip__grid li", func(_ int, el2 *colly.HTMLElement) {
	//					if el2.Attr("data-variantavailable") == "true" && el2.Attr("data-variantqty") != "0" {
	//						priceStr := el2.Attr("data-variantprice")
	//						priceStr = strings.Replace(priceStr, "$", "", -1)
	//						priceStr = strings.Replace(priceStr, ",", "", -1)
	//						priceStr = strings.Replace(priceStr, "SGD", "", -1)
	//						price, _ = strconv.ParseFloat(strings.TrimSpace(priceStr), 64)
	//						price = price / 100
	//
	//						// url with variant (quality)
	//						u := strings.TrimSpace(baseUrl + el.ChildAttr("a", "href"))
	//						cleanPageURL, err := url.Parse(u)
	//						if err != nil {
	//							log.Printf("error parsing url for %s with value [%s]: %v", storeName, u, err)
	//							return
	//						}
	//						cleanPageURL.RawQuery = url.Values{
	//							"variant":    []string{el2.Attr("data-variantid")},
	//							"utm_source": []string{config.UtmSource},
	//						}.Encode()
	//
	//						if price > 0 {
	//							cards = append(cards, gateway.Card{
	//								Name:      strings.TrimSpace(el.ChildText("p.productCard__title")),
	//								Url:       strings.TrimSpace(cleanPageURL.String()),
	//								InStock:   isInstock,
	//								Price:     price,
	//								Source:    storeName,
	//								Img:       strings.TrimSpace("https:" + el.ChildAttr("img", "data-src")),
	//								Quality:   el2.Attr("data-varianttitle"),
	//								ExtraInfo: []string{fmt.Sprintf("[%s]", el.ChildText("p.productCard__setName"))},
	//							})
	//						}
	//					}
	//				})
	//			}
	//		}
	//	})
	//})

	return cards, err
}

// card affinity
// cardboard crack games
// flagship games
// onemtg
// manapro
// mtgasia
func scrapVariant2(storeName, baseUrl, searchUrl, searchStr string) ([]gateway.Card, error) {
	searchURL := baseUrl + fmt.Sprintf(searchUrl, url.QueryEscape(searchStr+" mtg"))
	var cards []gateway.Card

	c := colly.NewCollector()

	c.OnHTML("body", func(e *colly.HTMLElement) {
		e.ForEach("div", func(_ int, el *colly.HTMLElement) {
			cardInfoStr := el.Attr("data-product-variants")
			if len(cardInfoStr) > 0 {
				productId := el.Attr("data-product-id")
				var pageUrl, imgUrl string
				if len(productId) > 0 {
					pageUrl = e.ChildAttr("div.product-card-list2__"+productId+" a", "href")
					imgUrl = e.ChildAttr("div.product-card-list2__"+productId+" img", "src")
				}

				var cardInfo []CardInfo
				err := json.Unmarshal([]byte(cardInfoStr), &cardInfo)
				if err == nil {
					if len(cardInfo) > 0 && len(pageUrl) > 0 && len(imgUrl) > 0 {
						for _, card := range cardInfo {
							// url with variant (quality)
							cleanPageURL, err := url.Parse(strings.TrimSpace(baseUrl + pageUrl))
							if err != nil {
								log.Printf("error parsing url for %s with value [%s]: %v", storeName, pageUrl, err)
								return
							}
							cleanPageURL.RawQuery = url.Values{
								"variant":    []string{fmt.Sprint(card.ID)},
								"utm_source": []string{config.UtmSource},
							}.Encode()

							cards = append(cards, gateway.Card{
								Name:    strings.TrimSpace(card.Name),
								Url:     strings.TrimSpace(cleanPageURL.String()),
								InStock: card.Available,
								Price:   float64(card.Price) / 100,
								Source:  storeName,
								Img:     strings.TrimSpace("https:" + imgUrl),
								Quality: card.Title,
							})
						}
					}
				}
			}
		})
	})

	return cards, c.Visit(searchURL)
}

// cards citadel
func scrapVariant1(storeName, baseUrl, searchUrl, searchStr string) ([]gateway.Card, error) {
	searchURL := baseUrl + fmt.Sprintf(searchUrl, url.QueryEscape(searchStr+" mtg"))
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

						u := strings.TrimSpace(baseUrl + el.ChildAttr("a", "href"))
						cleanPageURL, err := url.Parse(u)
						if err != nil {
							log.Printf("error parsing url for %s with value [%s]: %v", storeName, u, err)
							return
						}
						cleanPageURL.RawQuery = url.Values{
							"utm_source": []string{config.UtmSource},
						}.Encode()

						if price > 0 {
							cards = append(cards, gateway.Card{
								Name:    strings.TrimSpace(el.ChildText("p.productTitle")),
								Url:     strings.TrimSpace(cleanPageURL.String()),
								InStock: isInstock,
								Price:   price,
								Source:  storeName,
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
	priceStr = strings.Replace(priceStr, "From", "", -1)
	priceStr = strings.Replace(priceStr, "$", "", -1)
	priceStr = strings.Replace(priceStr, ",", "", -1)
	priceStr = strings.Replace(priceStr, "SGD", "", -1)
	priceStr = strings.TrimSpace(priceStr)
	return strconv.ParseFloat(priceStr, 64)
}
