package cardkingdom

import (
	"context"
	"fmt"
	"log"
)

const (
	ckPricelistErrorPrefixPublic = "ck price pricelist"
	mtgjsonErrorPrefix           = "ck price mtgjson"
)

// FetchCheapestByName downloads Card Kingdom retail prices and indexes the
// cheapest listed price per card name. It prefers the official CK pricelist API
// and falls back to MTGJSON when Cloudflare blocks the pricelist download.
func FetchCheapestByName(ctx context.Context) (map[string]Listing, error) {
	listings, err := fetchCheapestFromCKPricelist(ctx)
	if err == nil {
		return listings, nil
	}

	log.Printf("ck price refresh: card kingdom pricelist unavailable, falling back to mtgjson: %v", err)

	listings, fallbackErr := fetchCheapestFromMTGJSON(ctx)
	if fallbackErr != nil {
		return nil, fmt.Errorf("%s: %w; %s: %v", ckPricelistErrorPrefixPublic, err, mtgjsonErrorPrefix, fallbackErr)
	}
	return listings, nil
}
