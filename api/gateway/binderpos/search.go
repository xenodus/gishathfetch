package binderpos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/config"
)

const apiEndpoint = "https://portal.binderpos.com/external/shopify/products/forStore"
const productPath = "/products/"

// CardInfo is the structure of the card information from scrapping
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

func (i impl) Search(storeName, storeBaseURL string, payload []byte) ([]gateway.Card, int, error) {
	var cards []gateway.Card

	res, httpStatusCode, err := i.getApiResponse(payload)
	if err != nil {
		return cards, httpStatusCode, err
	}

	log.Printf("api response count for [%s]: [%d]", storeName, res.Count)

	if res.Count > 0 {
		for _, card := range res.Products {
			for _, stock := range card.Variants {
				if stock.Quantity > 0 {
					// url
					u := fmt.Sprintf("%s%s%s", storeBaseURL, productPath, card.Handle)
					cleanPageURL, err := url.Parse(strings.TrimSpace(u))
					if err != nil {
						log.Printf("error parsing url for %s with value [%s]: %v", storeName, u, err)
						continue
					}
					cleanPageURL.RawQuery = url.Values{
						"variant":    []string{fmt.Sprint(stock.ShopifyID)},
						"utm_source": []string{config.UtmSource},
					}.Encode()

					cards = append(cards, gateway.Card{
						Name:      fmt.Sprintf("%s", card.CardName),
						Url:       cleanPageURL.String(),
						InStock:   true,
						Price:     stock.Price,
						Source:    storeName,
						Img:       card.Img,
						Quality:   stock.Title,
						ExtraInfo: []string{fmt.Sprintf("[%s]", card.SetName)},
					})
				}
			}
		}
	}

	return cards, httpStatusCode, nil
}

func (i impl) getApiResponse(payload []byte) (Response, int, error) {
	var res Response

	resp, err := i.searchClient.Post(i.apiEndpoint, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return res, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, resp.StatusCode, err
	}

	log.Printf("api response status code: [%v]", resp.StatusCode)

	if err = json.Unmarshal(body, &res); err != nil {
		return res, resp.StatusCode, err
	}

	return res, resp.StatusCode, nil
}

type Payload struct {
	StoreURL    string `json:"storeUrl"`
	Game        string `json:"game"`
	Title       string `json:"title"`
	InstockOnly bool   `json:"instockOnly"`
}

type Response struct {
	Limit          int `json:"limit"`
	Offset         int `json:"offset"`
	CurrentFilters struct {
		ID                         interface{} `json:"id"`
		Name                       interface{} `json:"name"`
		Notes                      interface{} `json:"notes"`
		StoreURL                   string      `json:"storeUrl"`
		Game                       string      `json:"game"`
		Title                      string      `json:"title"`
		SetNames                   interface{} `json:"setNames"`
		Strict                     bool        `json:"strict"`
		InstockOnly                bool        `json:"instockOnly"`
		Colors                     interface{} `json:"colors"`
		Types                      interface{} `json:"types"`
		Rarities                   interface{} `json:"rarities"`
		MonsterTypes               interface{} `json:"monsterTypes"`
		PriceOverrideType          interface{} `json:"priceOverrideType"`
		PriceGreaterThan           interface{} `json:"priceGreaterThan"`
		PriceLessThan              interface{} `json:"priceLessThan"`
		OverallQuantityGreaterThan interface{} `json:"overallQuantityGreaterThan"`
		OverallQuantityLessThan    interface{} `json:"overallQuantityLessThan"`
		QuantityGreaterThan        interface{} `json:"quantityGreaterThan"`
		QuantityLessThan           interface{} `json:"quantityLessThan"`
		Tags                       interface{} `json:"tags"`
		Vendors                    interface{} `json:"vendors"`
		ProductTypes               interface{} `json:"productTypes"`
		SpecialTraits              interface{} `json:"specialTraits"`
		Eras                       interface{} `json:"eras"`
		FabClasses                 interface{} `json:"fabClasses"`
		Editions                   interface{} `json:"editions"`
		SubTypes                   interface{} `json:"subTypes"`
		GameCharacters             interface{} `json:"gameCharacters"`
		Finishes                   interface{} `json:"finishes"`
		Barcode                    interface{} `json:"barcode"`
		Sku                        interface{} `json:"sku"`
		Variants                   interface{} `json:"variants"`
		SortTypes                  []struct {
			Type  string `json:"type"`
			Asc   bool   `json:"asc"`
			Order int    `json:"order"`
		} `json:"sortTypes"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"currentFilters"`
	Count    int `json:"count"`
	Products []struct {
		ID       int `json:"id"`
		Variants []struct {
			ID                            int         `json:"id"`
			ShopifyID                     int64       `json:"shopifyId"`
			ProductTitle                  interface{} `json:"productTitle"`
			TcgImage                      interface{} `json:"tcgImage"`
			CollectorNumber               interface{} `json:"collectorNumber"`
			Img                           interface{} `json:"img"`
			Title                         string      `json:"title"`
			Barcode                       string      `json:"barcode"`
			Sku                           string      `json:"sku"`
			Price                         float64     `json:"price"`
			CashBuyPrice                  interface{} `json:"cashBuyPrice"`
			StoreCreditBuyPrice           interface{} `json:"storeCreditBuyPrice"`
			MaxPurchaseQuantity           interface{} `json:"maxPurchaseQuantity"`
			CanPurchaseOverstock          interface{} `json:"canPurchaseOverstock"`
			CreditOverstockBuyPrice       interface{} `json:"creditOverstockBuyPrice"`
			OvertStockBuyPrice            interface{} `json:"overtStockBuyPrice"`
			Quantity                      int         `json:"quantity"`
			ReserveQuantity               int         `json:"reserveQuantity"`
			Position                      int         `json:"position"`
			Taxable                       interface{} `json:"taxable"`
			TaxCode                       interface{} `json:"taxCode"`
			Option1                       string      `json:"option1"`
			Option2                       interface{} `json:"option2"`
			Option3                       interface{} `json:"option3"`
			FulfillmentService            string      `json:"fulfillmentService"`
			PriceOverride                 interface{} `json:"priceOverride"`
			CashBuyPercent                interface{} `json:"cashBuyPercent"`
			CreditBuyPercent              interface{} `json:"creditBuyPercent"`
			MaxInstockBuyPrice            interface{} `json:"maxInstockBuyPrice"`
			MaxInstockBuyPercentage       interface{} `json:"maxInstockBuyPercentage"`
			MaxInstockCreditBuyPrice      interface{} `json:"maxInstockCreditBuyPrice"`
			MaxInstockCreditBuyPercentage interface{} `json:"maxInstockCreditBuyPercentage"`
			VariantSyncSettings           interface{} `json:"variantSyncSettings"`
		} `json:"variants"`
		Event                          interface{} `json:"event"`
		ShopifyID                      int64       `json:"shopifyId"`
		SelectedVariant                int         `json:"selectedVariant"`
		OverallQuantity                int         `json:"overallQuantity"`
		Img                            string      `json:"img"`
		TcgImage                       string      `json:"tcgImage"`
		Title                          string      `json:"title"`
		Vendor                         string      `json:"vendor"`
		Tags                           string      `json:"tags"`
		Handle                         string      `json:"handle"`
		ProductType                    string      `json:"productType"`
		MetaFieldsGlobalDescriptionTag interface{} `json:"metaFieldsGlobalDescriptionTag"`
		MetaFieldsGlobalTitleTag       interface{} `json:"metaFieldsGlobalTitleTag"`
		TemplateSuffix                 interface{} `json:"templateSuffix"`
		Name                           interface{} `json:"name"`
		SetName                        string      `json:"setName"`
		SetCode                        string      `json:"setCode"`
		Rarity                         string      `json:"rarity"`
		CardName                       string      `json:"cardName"`
		CardTitle                      string      `json:"cardTitle"`
		CardNumber                     string      `json:"cardNumber"`
		CollectorNumber                string      `json:"collectorNumber"`
		ExtendedName                   interface{} `json:"extendedName"`
		SupportedCatalog               bool        `json:"supportedCatalog"`
	} `json:"products"`
}
