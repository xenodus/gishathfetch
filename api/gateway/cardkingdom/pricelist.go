package cardkingdom

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const (
	pricelistURL        = "https://api.cardkingdom.com/api/v2/pricelist"
	pricelistTimeout    = 3 * time.Minute
	outboundErrorPrefix = "ck price outbound"
)

var fetchPricelistBody = func(ctx context.Context) ([]byte, error) {
	body, err := fetchPricelistBodyTLS(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", outboundErrorPrefix, err)
	}
	return body, nil
}

// FetchCheapestByName downloads the Card Kingdom pricelist and indexes cheapest listings.
func FetchCheapestByName(ctx context.Context) (map[string]Listing, error) {
	body, err := fetchPricelistBody(ctx)
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
