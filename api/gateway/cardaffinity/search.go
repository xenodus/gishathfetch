package cardaffinity

import (
	"encoding/json"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/binderpos"
)

const StoreName = "Card Affinity"
const StoreBaseURL = "https://card-affinity.com"
const StoreSearchURL = "/search?q=%s"

const binderposStoreURL = "563304-2.myshopify.com"

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

	return binderpos.GetCards(s.Name, s.BaseUrl, reqPayload)
}

//type CardInfo struct {
//	ID                     int64    `json:"id"`
//	Title                  string   `json:"title"`
//	Option1                string   `json:"option1"`
//	Option2                any      `json:"option2"`
//	Option3                any      `json:"option3"`
//	Sku                    string   `json:"sku"`
//	RequiresShipping       bool     `json:"requires_shipping"`
//	Taxable                bool     `json:"taxable"`
//	FeaturedImage          any      `json:"featured_image"`
//	Available              bool     `json:"available"`
//	Name                   string   `json:"name"`
//	PublicTitle            string   `json:"public_title"`
//	Options                []string `json:"options"`
//	Price                  int      `json:"price"`
//	Weight                 int      `json:"weight"`
//	CompareAtPrice         any      `json:"compare_at_price"`
//	InventoryManagement    string   `json:"inventory_management"`
//	Barcode                any      `json:"barcode"`
//	RequiresSellingPlan    bool     `json:"requires_selling_plan"`
//	SellingPlanAllocations []any    `json:"selling_plan_allocations"`
//}

//func (s Store) Search(searchStr string) ([]scrapper.Card, error) {
//	searchURL := s.BaseUrl + fmt.Sprintf(s.SearchUrl, url.QueryEscape(searchStr))
//	var cards []scrapper.Card
//
//	c := colly.NewCollector()
//
//	c.OnHTML("body", func(e *colly.HTMLElement) {
//		e.ForEach("div", func(_ int, el *colly.HTMLElement) {
//			cardInfoStr := el.Attr("data-product-variants")
//			if len(cardInfoStr) > 0 {
//				productId := el.Attr("data-product-id")
//				var pageUrl, imgUrl string
//				if len(productId) > 0 {
//					pageUrl = e.ChildAttr("div.product-card-list2__"+productId+" a", "href")
//					imgUrl = e.ChildAttr("div.product-card-list2__"+productId+" img", "src")
//				}
//
//				var cardInfo []CardInfo
//				err := json.Unmarshal([]byte(cardInfoStr), &cardInfo)
//				if err == nil {
//					if len(cardInfo) > 0 && len(pageUrl) > 0 && len(imgUrl) > 0 {
//						for _, card := range cardInfo {
//							cards = append(cards, scrapper.Card{
//								Name:    strings.TrimSpace(card.Name),
//								Url:     strings.TrimSpace(s.BaseUrl + pageUrl),
//								InStock: card.Available,
//								Price:   float64(card.Price) / 100,
//								Source:  s.Name,
//								Img:     strings.TrimSpace("https:" + imgUrl),
//								Quality: card.Title,
//							})
//						}
//					}
//				}
//			}
//		})
//	})
//
//	return cards, c.Visit(searchURL)
//}
