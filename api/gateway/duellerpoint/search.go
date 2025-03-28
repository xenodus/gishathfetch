package duellerpoint

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"mtg-price-checker-sg/gateway"
)

const StoreName = "Dueller's Point"
const StoreBaseURL = "https://www.duellerspoint.com"
const StoreSearchURL = "/products/search?search_text=%s"

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

	doc.Find("div.container table > tbody").Each(func(i int, se *goquery.Selection) {
		se.Find("tr").Each(func(j int, se2 *goquery.Selection) {
			c := gateway.Card{
				Source: s.Name,
			}
			se2.Find("td").Each(func(k int, se3 *goquery.Selection) {
				switch k {
				case 0:
					c.Url = StoreBaseURL + se3.Find("a.product-list-thumb").AttrOr("href", "")
					c.Img = StoreBaseURL + se3.Find("a.product-list-thumb img").AttrOr("src", "")
				case 1:
					c.Name = strings.TrimSpace(se3.Text())
				case 2:
					c.Name += " [" + strings.TrimSpace(se3.Text()) + "]"
				case 3:
					se3.Find("p").Each(func(l int, se4 *goquery.Selection) {
						if strings.Contains(se4.Find("span").Text(), "Condition") {
							c.Quality = se4.Find("strong").Text()
						}
					})
				case 4:
					if strings.Contains(se3.Text(), "left") {
						c.InStock = true
					}
				case 5:
					price, err := parsePrice(se3.Text())
					if err != nil {
						break
					}
					c.Price = price
				}
			})
			if c.InStock {
				c.IsScrapped = true
				cards = append(cards, c)
			}
		})
	})

	return cards, nil
}

func parsePrice(price string) (float64, error) {
	priceStr := strings.TrimSpace(price)
	priceStr = strings.Replace(priceStr, "S$", "", -1)
	priceStr = strings.Replace(priceStr, ",", "", -1)
	return strconv.ParseFloat(priceStr, 64)
}
