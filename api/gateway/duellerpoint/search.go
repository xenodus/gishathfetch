package duellerpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
)

const StoreName = "Dueller's Point"
const StoreBaseURL = "https://www.duellerspoint.com"
const StoreSearchPath = "/products/search"
const tcgplayerImageURL = "https://product-images.tcgplayer.com/fit-in/437x437/%s.jpg"

type Store struct {
	Name       string
	BaseUrl    string
	SearchPath string
}

type searchResponse struct {
	Results []searchResult `json:"results"`
}

type searchResult struct {
	Name              string `json:"name"`
	Slug              string `json:"slug"`
	Price             string `json:"price"`
	FoilPrice         string `json:"foil_price"`
	Quantity          int    `json:"quantity"`
	IsActive          bool   `json:"is_active"`
	FoilType          string `json:"foil_type"`
	GetNameWithFoil   string `json:"get_name_with_foil"`
	GetVariationName  string `json:"get_variation_name"`
	TCGPlayerProductID string `json:"tcgplayer_product_id_ext"`
}

func NewLGS() gateway.LGS {
	return Store{
		Name:       StoreName,
		BaseUrl:    StoreBaseURL,
		SearchPath: StoreSearchPath,
	}
}

func (s Store) Search(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	apiURL := &url.URL{
		Scheme: "https",
		Host:   "www.duellerspoint.com",
		Path:   StoreSearchPath,
		RawQuery: url.Values{
			"search_text": {searchStr},
		}.Encode(),
	}

	resp, err := gateway.DoOutboundGET(ctx, apiURL.String(), gateway.OutboundRequestOptions{
		PageURL: apiURL,
		Accept:  "application/json",
	}, config.SearchAttemptTimeout)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := gateway.ReadResponseBody(resp)
	if err != nil {
		return nil, err
	}

	var payload searchResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, gateway.WrapJSONDecodeError(err, resp, body)
	}

	cards := make([]gateway.Card, 0, len(payload.Results))
	for _, item := range payload.Results {
		card, ok := parseSearchResult(item)
		if !ok {
			continue
		}

		cleanPageURL, err := url.Parse(card.Url)
		if err != nil {
			log.Printf("error parsing url for %s with value [%s]: %v", s.Name, card.Url, err)
			continue
		}
		cleanPageURL.RawQuery = url.Values{
			"utm_source": []string{config.UtmSource},
		}.Encode()
		card.Url = cleanPageURL.String()
		cards = append(cards, card)
	}

	return cards, nil
}

func parseSearchResult(item searchResult) (gateway.Card, bool) {
	if !item.IsActive || item.Quantity <= 0 || strings.TrimSpace(item.Slug) == "" {
		return gateway.Card{}, false
	}

	price, ok := listingPrice(item)
	if !ok {
		return gateway.Card{}, false
	}

	name := strings.TrimSpace(item.GetNameWithFoil)
	if name == "" {
		name = strings.TrimSpace(item.Name)
	}
	if name == "" {
		return gateway.Card{}, false
	}

	card := gateway.Card{
		Name:    name,
		Url:     StoreBaseURL + "/products/" + item.Slug,
		Img:     tcgplayerImageURLForProduct(item.TCGPlayerProductID),
		Price:   price,
		InStock: true,
		IsFoil:  item.FoilType == "foil" || strings.Contains(name, "Foil"),
		Source:  StoreName,
	}
	if setName := strings.TrimSpace(item.GetVariationName); setName != "" {
		card.ExtraInfo = []string{fmt.Sprintf("[%s]", setName)}
	}
	return card, card.Img != ""
}

func listingPrice(item searchResult) (float64, bool) {
	priceStr := strings.TrimSpace(item.Price)
	if item.FoilType == "foil" {
		if foilPrice := strings.TrimSpace(item.FoilPrice); foilPrice != "" && foilPrice != "0.0" && foilPrice != "0" {
			priceStr = foilPrice
		}
	}
	price, err := util.ParsePrice(priceStr)
	if err != nil || price <= 0 {
		return 0, false
	}
	return price, true
}

func tcgplayerImageURLForProduct(productIDExt string) string {
	productID := strings.TrimSpace(productIDExt)
	if productID == "" {
		return ""
	}
	if idx := strings.Index(productID, "-"); idx > 0 {
		productID = productID[:idx]
	}
	if _, err := strconv.Atoi(productID); err != nil {
		return ""
	}
	return fmt.Sprintf(tcgplayerImageURL, productID)
}
