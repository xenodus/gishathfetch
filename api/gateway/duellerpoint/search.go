package duellerpoint

import (
	"context"
	"fmt"
	"log"
	"mtg-price-checker-sg/gateway/util"
	"net/http"
	"net/url"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/config"

	"github.com/PuerkitoBio/goquery"
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

func (s Store) Search(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	var cards []gateway.Card

	// Build the request URL safely using net/url to avoid uncontrolled data in network requests.
	baseURL, err := url.Parse(s.BaseUrl)
	if err != nil {
		return cards, fmt.Errorf("invalid base URL: %w", err)
	}

	searchPath := fmt.Sprintf(s.SearchUrl, url.QueryEscape(searchStr))
	ref, err := url.Parse(searchPath)
	if err != nil {
		return cards, fmt.Errorf("invalid search path: %w", err)
	}

	apiURL := baseURL.ResolveReference(ref)

	// Validate that the resolved URL still points to the expected host.
	if apiURL.Host != baseURL.Host {
		return cards, fmt.Errorf("resolved URL host %q does not match expected host %q", apiURL.Host, baseURL.Host)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return cards, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return cards, err
	}
	defer resp.Body.Close()

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
					c.IsFoil = strings.Contains(c.Name, "Foil") // case sensitive
				case 2:
					c.ExtraInfo = []string{fmt.Sprintf("[%s]", strings.TrimSpace(se3.Text()))}
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
					price, err := util.ParsePrice(se3.Text())
					if err != nil {
						break
					}
					c.Price = price
				}
			})
			if c.InStock {
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

				cards = append(cards, c)
			}
		})
	})

	return cards, nil
}
