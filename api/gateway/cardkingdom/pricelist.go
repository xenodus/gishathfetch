package cardkingdom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"mtg-price-checker-sg/gateway"
)

const pricelistURL = "https://api.cardkingdom.com/api/v2/pricelist"

var httpGet = func(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if err := gateway.PrepareOutboundRequest(ctx, req, gateway.OutboundRequestOptions{
		Style: gateway.OutboundStyleJSON,
	}); err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

// FetchCheapestByName downloads the Card Kingdom pricelist and indexes cheapest listings.
func FetchCheapestByName(ctx context.Context) (map[string]Listing, error) {
	resp, err := httpGet(ctx, pricelistURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cardkingdom: pricelist status %d", resp.StatusCode)
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
