package tcgmarketplace

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
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

type listing struct {
	ID                    int64  `json:"id"`
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
	if config.TCGMarketplaceAdvancedFilterSearchEnabled() {
		return s.SearchAdvancedFilter(ctx, searchStr)
	}

	var (
		listings    []listing
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

	listings, err = getApiResponse(ctx, reqPayload, accessToken != "")
	if err != nil {
		return cards, err
	}

	for _, listing := range listings {
		card, ok := cardFromListing(s, listing)
		if !ok {
			continue
		}
		cards = append(cards, card)
	}
	return cards, nil
}

// canonicalProductURL maps API product links onto the public storefront host.
// Post-maintenance responses sometimes use thetcgmarketplace.cc; the site serves on .com.
func canonicalProductURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(u.Host, "thetcgmarketplace.cc") {
		u.Host = "thetcgmarketplace.com"
	}
	return u, nil
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

func listingsFromEnvelope(env apiEnvelope) ([]listing, error) {
	if env.Status != http.StatusOK {
		msg := strings.TrimSpace(env.Data.Message)
		if msg == "" {
			return nil, fmt.Errorf("%s: search api status %d", StoreName, env.Status)
		}
		return nil, fmt.Errorf("%s: %s", StoreName, msg)
	}
	return decodeListings(env.Data.Data)
}

func decodeListings(raw json.RawMessage) ([]listing, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) || bytes.Equal(raw, []byte(`""`)) {
		return nil, nil
	}

	var listings []listing
	if err := json.Unmarshal(raw, &listings); err == nil {
		return listings, nil
	}

	var apiErr struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &apiErr); err == nil && (apiErr.Code != "" || apiErr.Message != "") {
		if apiErr.Code != "" {
			return nil, fmt.Errorf("%s: %s", StoreName, apiErr.Code)
		}
		return nil, fmt.Errorf("%s: %s", StoreName, apiErr.Message)
	}

	return nil, fmt.Errorf("%s: unexpected search response payload", StoreName)
}

func tcgMarketplaceOutboundOpts() gateway.OutboundRequestOptions {
	return gateway.OutboundRequestOptions{
		DirectAttemptTimeout:    config.TCGMarketplaceSearchAttemptTimeout,
		DedicatedAttemptTimeout: config.TCGMarketplaceSearchAttemptTimeout,
	}
}

func getApiResponse(ctx context.Context, payload []byte, accessTokenConfigured bool) ([]listing, error) {
	var requestContext []string
	if !accessTokenConfigured {
		requestContext = append(requestContext, "access_token_configured=false")
	}

	resp, err := gateway.DoOutboundRoundTrip(ctx, tcgMarketplaceOutboundOpts(), config.TCGMarketplaceSearchAttemptTimeout, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, cardLinkAPI, bytes.NewBuffer(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(payload))
		return req, nil
	})
	if err != nil {
		return nil, gateway.WrapHTTPRequestError(err, nil, requestContext...)
	}
	defer resp.Body.Close()

	body, err := gateway.ReadResponseBody(resp)
	if err != nil {
		return nil, gateway.WrapResponseBodyReadError(err, resp)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%s", gateway.FormatUnexpectedHTTPStatus(StoreName, resp, body))
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, gateway.WrapJSONDecodeError(err, resp, body)
	}
	listings, err := listingsFromEnvelope(env)
	if err != nil {
		return nil, err
	}

	return listings, nil
}
