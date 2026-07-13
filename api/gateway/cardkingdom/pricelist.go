package cardkingdom

import (
	"context"
	"fmt"
)

const mtgjsonErrorPrefix = "ck price mtgjson"

// FetchCheapestByName downloads Card Kingdom retail prices from MTGJSON and
// indexes the cheapest listing per card name.
func FetchCheapestByName(ctx context.Context) (map[string]Listing, error) {
	listings, err := fetchCheapestFromMTGJSON(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", mtgjsonErrorPrefix, err)
	}
	return listings, nil
}
