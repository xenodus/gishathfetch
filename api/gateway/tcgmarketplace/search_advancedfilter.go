package tcgmarketplace

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/config"
)

const advancedFilterAPI = "https://thetcgmarketplace.com:3501/product/advancedfilter"

type advancedFilterPayload struct {
	CategoryID     string `json:"category_id"`
	NameExactMatch bool   `json:"name_exact_match"`
	AvailableOnly  bool   `json:"available_only"`
	Name           string `json:"name"`
	Order          string `json:"order"`
}

// SearchAdvancedFilter queries POST /product/advancedfilter (no access token).
// It is an alternative to encoder/advancedsearch and is not wired into production search yet.
func (s Store) SearchAdvancedFilter(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	reqPayload, err := json.Marshal(advancedFilterPayload{
		CategoryID:     strconv.Itoa(mtgCategoryNo),
		NameExactMatch: false,
		AvailableOnly:  true,
		Name:           searchStr,
		Order:          "price_asc",
	})
	if err != nil {
		return nil, err
	}

	listings, err := getAdvancedFilterResponse(ctx, reqPayload)
	if err != nil {
		return nil, err
	}

	var cards []gateway.Card
	for _, listing := range listings {
		card, ok := cardFromListing(s, listing)
		if !ok {
			continue
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func getAdvancedFilterResponse(ctx context.Context, payload []byte) ([]listing, error) {
	resp, err := gateway.DoOutboundRoundTrip(ctx, tcgMarketplaceOutboundOpts(), config.TCGMarketplaceSearchAttemptTimeout, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, advancedFilterAPI, bytes.NewBuffer(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Referer", StoreBaseURL+"/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/153.0.0.0 Safari/537.36")
		req.ContentLength = int64(len(payload))
		return req, nil
	})
	if err != nil {
		return nil, gateway.WrapHTTPRequestError(err, nil)
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
	return listingsFromEnvelope(env)
}
