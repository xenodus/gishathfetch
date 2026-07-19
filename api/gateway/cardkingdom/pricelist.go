package cardkingdom

import "context"

// FetchCheapestByName downloads Card Kingdom retail prices from the official CK
// pricelist API and indexes the cheapest listed price per card name.
func FetchCheapestByName(ctx context.Context) (map[string]Listing, error) {
	return fetchCheapestFromCKPricelist(ctx)
}
