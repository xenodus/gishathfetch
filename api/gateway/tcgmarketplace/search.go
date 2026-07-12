package tcgmarketplace

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/config"
)

const StoreName = "The TCG Marketplace"
const StoreBaseURL = "https://thetcgmarketplace.com"

const cardLinkAPI = "https://thetcgmarketplace.com:3501/encoder/advancedsearch"
const mtgCategoryNo = 3
const accessTokenKey = "TCG_MARKETPLACE_ACCESS_TOKEN"

type response struct {
	Status int `json:"status"`
	Data   struct {
		Message string `json:"message"`
		Data    []struct {
			Name                  string `json:"name"`
			Setcode               string `json:"setcode"`
			Setname               string `json:"setname"`
			Image                 string `json:"image"`
			Language              string `json:"language"`
			CrdFoilType           any    `json:"crd_foil_type"`
			Rarity                string `json:"rarity"`
			Available             any    `json:"available"`
			From                  any    `json:"from"`
			NonFoilReferencePrice any    `json:"non_foil_reference_price"`
			FoilReferencePrice    any    `json:"foil_reference_price"`
			URL                   string `json:"url"`
		} `json:"data"`
	} `json:"data"`
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
}

type Store struct {
	Name      string
	BaseUrl   string
	SearchUrl string
}

type payload struct {
	AccessToken string `json:"access_token"`
	Name        string `json:"name"`
	Category    int32  `json:"category"`
	Order       string `json:"order"`
}

func NewLGS() gateway.LGS {
	return Store{
		Name:    StoreName,
		BaseUrl: StoreBaseURL,
	}
}

func (s Store) Search(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	var (
		res         response
		cards       []gateway.Card
		accessToken string
	)

	accessToken = os.Getenv(accessTokenKey)

	reqPayload, err := json.Marshal(payload{
		AccessToken: accessToken,
		Name:        searchStr,
		Category:    mtgCategoryNo,
		Order:       "name_asc",
	})
	if err != nil {
		return cards, err
	}

	res, err = getApiResponse(ctx, reqPayload, accessToken != "")
	if err != nil {
		return cards, err
	}

	if len(res.Data.Data) > 0 {
		for _, card := range res.Data.Data {
			stock, err := strconv.ParseInt(fmt.Sprint(card.Available), 10, 64)
			if err != nil {
				continue
			}

			if stock > 0 {
				price, err := strconv.ParseFloat(fmt.Sprint(card.From), 64)
				if err != nil {
					continue
				}

				// Strip [XXX] prefix from card name
				// e.g. [CMM] Deflecting Swat (V2)(Etched foil)
				name := strings.TrimSpace(card.Name)
				squareBracketIndex := strings.Index(name, "]")
				if squareBracketIndex > 1 {
					name = strings.TrimSpace(name[squareBracketIndex+1:])
				}

				var img string
				images := strings.Split(card.Image, " ")
				if len(images) > 0 {
					img = images[0]
				}

				// url
				u := strings.TrimSpace(card.URL)
				cleanPageURL, err := url.Parse(u)
				if err != nil {
					log.Printf("error parsing url for %s with value [%s]: %v", s.Name, u, err)
					continue
				}
				cleanPageURL.RawQuery = url.Values{
					"utm_source": []string{config.UtmSource},
				}.Encode()

				extraInfo := []string{fmt.Sprintf("[%s]", card.Setname)}
				cards = append(cards, gateway.Card{
					Name:      strings.TrimSpace(name),
					Url:       cleanPageURL.String(),
					InStock:   true,
					Price:     price,
					Source:    s.Name,
					Img:       img,
					IsFoil:    isSurgeFoil(extraInfo, name),
					ExtraInfo: extraInfo,
				})
			}
		}
	}
	return cards, nil
}

func isSurgeFoil(extraInfo []string, name string) bool {
	if strings.Contains(name, "Surge Foil") {
		return true
	}
	for _, info := range extraInfo {
		if strings.Contains(info, "Surge Foil") {
			return true
		}
	}
	return false
}

func getApiResponse(ctx context.Context, payload []byte, accessTokenConfigured bool) (response, error) {
	var res response

	var requestContext []string
	if !accessTokenConfigured {
		requestContext = append(requestContext, "access_token_configured=false")
	}

	resp, err := gateway.DoOutboundRoundTrip(ctx, gateway.OutboundRequestOptions{}, config.SearchAttemptTimeout, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, cardLinkAPI, bytes.NewBuffer(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(payload))
		return req, nil
	})
	if err != nil {
		return res, gateway.WrapHTTPRequestError(err, nil, requestContext...)
	}
	defer resp.Body.Close()

	body, err := gateway.ReadResponseBody(resp)
	if err != nil {
		return res, gateway.WrapResponseBodyReadError(err, resp)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return res, fmt.Errorf("%s", gateway.FormatUnexpectedHTTPStatus(StoreName, resp, body))
	}
	if err = json.Unmarshal(body, &res); err != nil {
		return res, gateway.WrapJSONDecodeError(err, resp, body)
	}

	return res, nil
}
