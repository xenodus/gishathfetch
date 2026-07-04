package tcgmarketplace

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

type apiEnvelope struct {
	Status int `json:"status"`
	Data   struct {
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	} `json:"data"`
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
}

type cardItem struct {
	Name                  string      `json:"name"`
	Setcode               string      `json:"setcode"`
	Setname               string      `json:"setname"`
	Image                 string      `json:"image"`
	Language              string      `json:"language"`
	CrdFoilType           interface{} `json:"crd_foil_type"`
	Rarity                string      `json:"rarity"`
	Available             interface{} `json:"available"`
	From                  interface{} `json:"from"`
	NonFoilReferencePrice interface{} `json:"non_foil_reference_price"`
	FoilReferencePrice    interface{} `json:"foil_reference_price"`
	URL                   string      `json:"url"`
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

	items, err := getApiResponse(ctx, reqPayload, accessToken != "")
	if err != nil {
		return cards, err
	}

	if len(items) > 0 {
		for _, card := range items {
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

func parseResponseBody(body []byte) ([]cardItem, error) {
	var envelope apiEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}

	if envelope.Status != http.StatusOK {
		return nil, apiError(envelope.Status, envelope.Data.Message, envelope.Data.Data)
	}

	return parseCardItems(envelope.Data.Data)
}

func parseCardItems(raw json.RawMessage) ([]cardItem, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == `""` || trimmed == "null" {
		return nil, nil
	}
	if !strings.HasPrefix(trimmed, "[") {
		return nil, fmt.Errorf("%s API returned unexpected data payload: %s", StoreName, trimmed)
	}

	var cards []cardItem
	if err := json.Unmarshal(raw, &cards); err != nil {
		return nil, err
	}
	return cards, nil
}

func apiError(status int, message string, raw json.RawMessage) error {
	detail := strings.TrimSpace(message)
	if detail == "" {
		detail = describeErrorPayload(raw)
	}
	if detail == "" {
		detail = "unknown error"
	}
	return fmt.Errorf("%s API error (status=%d): %s", StoreName, status, detail)
}

func describeErrorPayload(raw json.RawMessage) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == `""` || trimmed == "null" {
		return ""
	}
	if !strings.HasPrefix(trimmed, "{") {
		return trimmed
	}

	var errData struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Syscall string `json:"syscall"`
		Address string `json:"address"`
		Port    int    `json:"port"`
	}
	if err := json.Unmarshal(raw, &errData); err != nil {
		return trimmed
	}

	switch {
	case errData.Code != "" && errData.Address != "":
		return fmt.Sprintf("%s (%s:%d)", errData.Code, errData.Address, errData.Port)
	case errData.Code != "":
		return errData.Code
	case errData.Message != "":
		return errData.Message
	default:
		return trimmed
	}
}

func getApiResponse(ctx context.Context, payload []byte, accessTokenConfigured bool) ([]cardItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cardLinkAPI, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(payload))

	var requestContext []string
	if !accessTokenConfigured {
		requestContext = append(requestContext, "access_token_configured=false")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, gateway.WrapHTTPRequestError(err, req, requestContext...)
	}
	defer resp.Body.Close()

	body, err := gateway.ReadResponseBody(resp)
	if err != nil {
		return nil, gateway.WrapResponseBodyReadError(err, resp)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%s", gateway.FormatUnexpectedHTTPStatus(StoreName, resp, body))
	}

	items, err := parseResponseBody(body)
	if err != nil {
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			return nil, gateway.WrapJSONDecodeError(err, resp, body)
		}
		return nil, err
	}

	return items, nil
}
