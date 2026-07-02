package cardkingdom

import (
	"context"
	"fmt"
)

const mtgjsonErrorPrefix = "ck price mtgjson"

// FetchCheapestByName downloads Card Kingdom retail prices from MTGJSON and
// indexes the cheapest listing per card name. When available, Card Kingdom's
// pricelist is merged in as well so face-only names (for example "Jennifer
// Walters") supplement MTGJSON's full double-faced names.
func FetchCheapestByName(ctx context.Context) (map[string]Listing, error) {
	listings, err := fetchCheapestFromMTGJSON(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", mtgjsonErrorPrefix, err)
	}

	if ckListings, ckErr := fetchCheapestFromCKPricelist(ctx); ckErr == nil {
		mergeCheapestListings(listings, ckListings)
	}

	return listings, nil
}
