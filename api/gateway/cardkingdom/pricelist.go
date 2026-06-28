package cardkingdom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"mtg-price-checker-sg/gateway"
)

const (
	pricelistURL          = "https://api.cardkingdom.com/api/v2/pricelist"
	pricelistTimeout      = 3 * time.Minute
	outboundErrorPrefix   = "ck price outbound"
)

var fetchPricelistResponse = func(ctx context.Context) (*http.Response, error) {
	storeBase, err := url.Parse(listingBaseURL)
	if err != nil {
		return nil, err
	}

	return gateway.DoOutboundGET(ctx, pricelistURL, gateway.OutboundRequestOptions{
		Style:          gateway.OutboundStyleJSON,
		StoreBase:      storeBase,
		SkipWebBotAuth: true,
	}, pricelistTimeout)
}

// FetchCheapestByName downloads the Card Kingdom pricelist and indexes cheapest listings.
func FetchCheapestByName(ctx context.Context) (map[string]Listing, error) {
	resp, err := fetchPricelistResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", outboundErrorPrefix, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: pricelist status %d", outboundErrorPrefix, resp.StatusCode)
	}

	body, err := gateway.ReadResponseBody(resp)
	if err != nil {
		return nil, err
	}

	var products []Product
	if err := json.Unmarshal(body, &products); err == nil {
		return BuildCheapestByName(products, time.Now().UTC()), nil
	}

	var payload struct {
		Data []Product `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	return BuildCheapestByName(payload.Data, time.Now().UTC()), nil
}
