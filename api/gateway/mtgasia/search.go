package mtgasia

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/binderpos"
)

const StoreName = "MTG Asia"
const StoreBaseURL = "https://www.mtg-asia.com"
const StoreSearchURL = "/search?q=%s"

const binderposStoreURL = "mtgasia.myshopify.com"

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

type CardInfo struct {
	ID                     int64    `json:"id"`
	Title                  string   `json:"title"`
	Option1                string   `json:"option1"`
	Option2                any      `json:"option2"`
	Option3                any      `json:"option3"`
	Sku                    string   `json:"sku"`
	RequiresShipping       bool     `json:"requires_shipping"`
	Taxable                bool     `json:"taxable"`
	FeaturedImage          any      `json:"featured_image"`
	Available              bool     `json:"available"`
	Name                   string   `json:"name"`
	PublicTitle            string   `json:"public_title"`
	Options                []string `json:"options"`
	Price                  int      `json:"price"`
	Weight                 int      `json:"weight"`
	CompareAtPrice         any      `json:"compare_at_price"`
	InventoryManagement    string   `json:"inventory_management"`
	Barcode                any      `json:"barcode"`
	RequiresSellingPlan    bool     `json:"requires_selling_plan"`
	SellingPlanAllocations []any    `json:"selling_plan_allocations"`
}

func scrap(s Store, searchStr string) ([]gateway.Card, error) {
	searchURL := s.BaseUrl + fmt.Sprintf(s.SearchUrl, url.QueryEscape(searchStr))
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
							cards = append(cards, gateway.Card{
								Name:       strings.TrimSpace(card.Name),
								Url:        strings.TrimSpace(s.BaseUrl + pageUrl),
								InStock:    card.Available,
								Price:      float64(card.Price) / 100,
								Source:     s.Name,
								Img:        strings.TrimSpace("https:" + imgUrl),
								Quality:    card.Title,
								IsScrapped: true,
							})
						}
					}
				}
			}
		})
	})

	return cards, c.Visit(searchURL)
}
