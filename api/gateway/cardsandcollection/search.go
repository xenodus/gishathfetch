package cardsandcollection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/config"
)

const StoreName = "Cards & Collections"
const StoreBaseURL = "https://cardsandcollections.com"
const StoreApiURL = "/api/catalog/"
const StoreSearchURL = "/?q="

type response struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total    int `json:"total"`
		MaxScore any `json:"max_score"`
		Hits     []struct {
			Index  string `json:"_index"`
			Type   string `json:"_type"`
			ID     string `json:"_id"`
			Score  any    `json:"_score"`
			Source struct {
				ID                          string  `json:"id"`
				Name                        string  `json:"name"`
				OriginalName                string  `json:"originalName"`
				CollectableContext          string  `json:"collectableContext"`
				IsProductSingle             bool    `json:"isProductSingle"`
				ProductCategory             string  `json:"productCategory"`
				Artist                      string  `json:"artist"`
				ASCIIName                   any     `json:"asciiName"`
				BorderColor                 string  `json:"borderColor"`
				ColorIdentity               string  `json:"colorIdentity"`
				ColorIndicator              any     `json:"colorIndicator"`
				Colors                      string  `json:"colors"`
				ConvertedManaCost           string  `json:"convertedManaCost"`
				FaceConvertedManaCost       any     `json:"faceConvertedManaCost"`
				FaceName                    any     `json:"faceName"`
				FlavorName                  any     `json:"flavorName"`
				FlavorText                  string  `json:"flavorText"`
				FrameEffects                any     `json:"frameEffects"`
				FrameVersion                string  `json:"frameVersion"`
				HasFoil                     bool    `json:"hasFoil"`
				HasNonFoil                  bool    `json:"hasNonFoil"`
				IsAlternative               bool    `json:"isAlternative"`
				IsFullArt                   bool    `json:"isFullArt"`
				IsOversized                 bool    `json:"isOversized"`
				IsPromo                     bool    `json:"isPromo"`
				IsReprint                   bool    `json:"isReprint"`
				IsReserved                  bool    `json:"isReserved"`
				IsStarter                   bool    `json:"isStarter"`
				IsStorySpotlight            bool    `json:"isStorySpotlight"`
				IsTextless                  bool    `json:"isTextless"`
				IsTimeshifted               bool    `json:"isTimeshifted"`
				Keywords                    any     `json:"keywords"`
				Layout                      string  `json:"layout"`
				LeadershipSkills            any     `json:"leadershipSkills"`
				Life                        any     `json:"life"`
				Loyalty                     any     `json:"loyalty"`
				ManaCost                    string  `json:"manaCost"`
				MultiverseID                string  `json:"multiverseId"`
				Img                         string  `json:"img"`
				OtherFaceImg                string  `json:"otherFaceImg"`
				CropArtImg                  string  `json:"cropArtImg"`
				Number                      string  `json:"number"`
				OriginalText                string  `json:"originalText"`
				OriginalType                string  `json:"originalType"`
				OtherFaceIds                any     `json:"otherFaceIds"`
				Power                       any     `json:"power"`
				Printings                   string  `json:"printings"`
				PromoTypes                  any     `json:"promoTypes"`
				Rarity                      string  `json:"rarity"`
				KeyruneCode                 string  `json:"keyruneCode"`
				SetCode                     string  `json:"setCode"`
				SetName                     string  `json:"setName"`
				SetNameForFilter            string  `json:"setNameForFilter"`
				Side                        any     `json:"side"`
				Subtypes                    string  `json:"subtypes"`
				Supertypes                  string  `json:"supertypes"`
				TcgplayerProductID          string  `json:"tcgplayerProductId"`
				Text                        string  `json:"text"`
				Toughness                   any     `json:"toughness"`
				Type                        string  `json:"type"`
				Types                       string  `json:"types"`
				UUID                        string  `json:"uuid"`
				Variations                  string  `json:"variations"`
				AvgPriceSale                any     `json:"avgPriceSale"`
				AvgPriceBuy                 any     `json:"avgPriceBuy"`
				MinPriceSale                any     `json:"minPriceSale"`
				MinPriceBuy                 any     `json:"minPriceBuy"`
				QuantityOnSale              any     `json:"quantityOnSale"`
				QuantityOnBuy               any     `json:"quantityOnBuy"`
				SaleItemUpdatedAtOnSale     any     `json:"saleItemUpdatedAtOnSale"`
				SaleItemUpdatedAtOnBuy      any     `json:"saleItemUpdatedAtOnBuy"`
				SealedProductCategory       any     `json:"sealedProductCategory"`
				SealedProductSubtype        any     `json:"sealedProductSubtype"`
				AccessoryProductType        any     `json:"AccessoryProductType"`
				AccessoryProductBrand       any     `json:"AccessoryProductBrand"`
				AccessoryProductDescription any     `json:"AccessoryProductDescription"`
				Willpower                   any     `json:"willpower"`
				Strength                    any     `json:"strength"`
				Abilities                   any     `json:"abilities"`
				Classifications             any     `json:"classifications"`
				Cost                        any     `json:"cost"`
				Lore                        any     `json:"lore"`
				CardkingdomPrice            float64 `json:"cardkingdomPrice"`
				CardkingdomUpdatedAt        string  `json:"cardkingdomUpdatedAt"`
			} `json:"_source"`
			Sort []int64 `json:"sort"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations struct {
		ProductCategory4 struct {
			DocCount           int `json:"doc_count"`
			ProductCategoryRaw struct {
				DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
				SumOtherDocCount        int `json:"sum_other_doc_count"`
				Buckets                 []struct {
					Key      string `json:"key"`
					DocCount int    `json:"doc_count"`
				} `json:"buckets"`
			} `json:"productCategory.raw"`
			ProductCategoryRawCount struct {
				Value int `json:"value"`
			} `json:"productCategory.raw_count"`
		} `json:"productCategory4"`
	} `json:"aggregations"`
}

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
	var (
		res   response
		cards []gateway.Card
	)

	apiURL := s.BaseUrl + StoreApiURL
	requestBody := []byte(fmt.Sprintf(`{"query":{"bool":{"should":[{"simple_query_string":{"query":"%s","fields":["name","setCode","setName"],"default_operator":"AND"}},{"multi_match":{"query":"%s","type":"phrase_prefix","fields":["name","setCode","setName"]}}]}},"post_filter":{"bool":{"must":{"terms":{"collectableContext.raw":["MTG","ACCESSORY"]}}}},"aggs":{"productCategory4":{"filter":{"bool":{"must":{"terms":{"collectableContext.raw":["MTG","ACCESSORY"]}}}},"aggs":{"productCategory.raw":{"terms":{"field":"productCategory.raw","size":50}},"productCategory.raw_count":{"cardinality":{"field":"productCategory.raw"}}}}},"size":20,"sort":[{"quantityOnSale":"desc"}]}`, searchStr, searchStr))
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return cards, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return cards, err
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return cards, err
	}

	if res.Hits.Total > 0 {
		for _, hit := range res.Hits.Hits {
			if quantityOnSaleStr, quantityOk := hit.Source.QuantityOnSale.(string); quantityOk {
				if minPriceStr, priceOk := hit.Source.MinPriceSale.(string); priceOk {
					quantity, _ := strconv.Atoi(quantityOnSaleStr)
					minPrice, _ := strconv.ParseFloat(minPriceStr, 64)

					if quantity > 0 && minPrice > 0 {
						// url
						u := fmt.Sprintf(StoreBaseURL+"/product/%s", hit.ID)
						cleanPageURL, err := url.Parse(u)
						if err != nil {
							log.Printf("error parsing url for %s with value [%s]: %v", s.Name, u, err)
							continue
						}
						cleanPageURL.RawQuery = url.Values{
							"utm_source": []string{config.UtmSource},
						}.Encode()

						cards = append(cards, gateway.Card{
							Name:      strings.TrimSpace(hit.Source.Name),
							Url:       strings.TrimSpace(cleanPageURL.String()),
							InStock:   true,
							Price:     minPrice,
							Source:    s.Name,
							Img:       hit.Source.Img,
							ExtraInfo: []string{fmt.Sprintf("[%s]", hit.Source.SetName)},
						})
					}
				}
			}
		}
	}

	return cards, nil
}
