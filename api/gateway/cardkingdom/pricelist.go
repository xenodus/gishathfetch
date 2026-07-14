package cardkingdom

import (
	"context"
	"fmt"
)

const ckPricelistErrorPrefixPublic = "ck price pricelist"

// FetchCheapestByName downloads Card Kingdom's public pricelist and indexes the
// cheapest in-stock retail listing per card name.
func FetchCheapestByName(ctx context.Context) (map[string]Listing, error) {
	listings, err := fetchCheapestFromCKPricelist(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ckPricelistErrorPrefixPublic, err)
	}
	return listings, nil
}
