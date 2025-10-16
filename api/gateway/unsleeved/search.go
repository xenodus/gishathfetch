package unsleeved

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const StoreName = "Unsleeved"
const StoreBaseURL = "https://hitpay.shop/unsleeved"
const StoreSearchURL = "/search"
const StoreApiURL = "https://hitpay.shop/api/v1/search-products?keywords=%s"
const StoreHeaderIDKey = "hitpay-identifier"
const StoreHeaderIDVal = "9879a70f-ebff-4612-9818-2ad353f94dee"

type Store struct {
	Name      string
	BaseUrl   string
	SearchUrl string
}

type response struct {
	Data []struct {
		ID         string `json:"id"`
		BusinessID string `json:"business_id"`
		// Skipping CategoryID
		Name                   string      `json:"name"`
		Headline               interface{} `json:"headline"`
		Description            string      `json:"description"`
		StockKeepingUnit       string      `json:"stock_keeping_unit"`
		Currency               string      `json:"currency"`
		Price                  float64     `json:"price"`
		PriceBeforeDiscount    interface{} `json:"price_before_discount"`
		PriceDisplay           string      `json:"price_display"`
		PriceStored            int         `json:"price_stored"`
		IsManageable           int         `json:"is_manageable"`
		IsPinned               bool        `json:"is_pinned"`
		Status                 string      `json:"status"`
		ProductWeight          int         `json:"product_weight"`
		DeliveryMethodRequired bool        `json:"delivery_method_required"`
		HasVariations          bool        `json:"has_variations"`
		IsShopify              bool        `json:"is_shopify"`
		IsWoocommerce          bool        `json:"is_woocommerce"`
		Order                  int         `json:"order"`
		Quantity               int         `json:"quantity"`
		QuantityAlertLevel     interface{} `json:"quantity_alert_level"`
		MinOrderQuantity       interface{} `json:"min_order_quantity"`
		MaxOrderQuantity       interface{} `json:"max_order_quantity"`
		Emoji                  interface{} `json:"emoji"`
		OpenAmount             bool        `json:"open_amount"`
		ProductURL             string      `json:"product_url"` // for linking to product page
		VariationsCount        int         `json:"variations_count"`
		// Skipping Images, using single dimension from field below
		Image               string      `json:"image"`
		IsPublished         bool        `json:"is_published"`
		PublishedAt         interface{} `json:"published_at"`
		CreatedAt           time.Time   `json:"created_at"`
		UpdatedAt           time.Time   `json:"updated_at"`
		OrderInCategory     interface{} `json:"order_in_category"`
		AllowBackOrder      bool        `json:"allow_back_order"`
		Available           bool        `json:"available"`
		Type                string      `json:"type"`
		PasswordProtected   bool        `json:"password_protected"`
		DigitalContent      interface{} `json:"digital_content"`
		AutoTagNewLocations bool        `json:"auto_tag_new_locations"`
		Channels            []string    `json:"channels"`
		// Skipping Locations
		ProductUnit                   interface{}   `json:"product_unit"`
		ProductUnitAbbreviation       interface{}   `json:"product_unit_abbreviation"`
		ProductUnitValue              int           `json:"product_unit_value"`
		Handle                        string        `json:"handle"`
		PosColor                      interface{}   `json:"pos_color"`
		ProductAddOns                 []interface{} `json:"product_add_ons"`
		IsInventoryTracked            bool          `json:"is_inventory_tracked"`
		IsOnlineStoreInventoryTracked bool          `json:"is_online_store_inventory_tracked"`
	} `json:"data"`
	// Skipping Meta & Links (for pagination)
}

func NewLGS() gateway.LGS {
	return Store{
		Name:      StoreName,
		BaseUrl:   StoreBaseURL,
		SearchUrl: StoreSearchURL,
	}
}

func (s Store) Search(searchStr string) ([]gateway.Card, error) {
	var (
		res   response
		cards []gateway.Card
	)

	apiURL := fmt.Sprintf(StoreApiURL, url.QueryEscape(searchStr))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return cards, err
	}

	// Set custom headers
	req.Header.Set(StoreHeaderIDKey, StoreHeaderIDVal)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return cards, err
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return cards, err
	}

	if len(res.Data) > 0 {
		for _, card := range res.Data {
			price, err := util.ParsePrice(card.PriceDisplay)
			if err != nil {
				log.Printf("error parsing price for %s with value [%s]: %v", s.Name, card.PriceDisplay, err)
				return cards, err
			}

			if card.Status == "published" && card.Quantity > 0 {
				u := fmt.Sprintf("%s/product/%s", StoreBaseURL, card.Handle)
				cleanPageURL, err := url.Parse(u)
				if err != nil {
					log.Printf("error parsing url for %s with value [%s]: %v", s.Name, u, err)
					return cards, err
				}
				cleanPageURL.RawQuery = url.Values{
					"utm_source": []string{config.UtmSource},
				}.Encode()

				cards = append(cards, gateway.Card{
					Name:      strings.TrimSpace(card.Name),
					Url:       cleanPageURL.String(),
					InStock:   true,
					Price:     price,
					Source:    s.Name,
					Img:       card.Image,
					ExtraInfo: []string{card.Description},
				})
			}
		}
	}

	return cards, nil
}
