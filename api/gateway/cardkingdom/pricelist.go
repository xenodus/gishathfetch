package cardkingdom

import "context"

// PricelistFetchResult is the outcome of downloading and indexing the CK pricelist.
type PricelistFetchResult struct {
	Listings       map[string]Listing
	TransportOrder string
}

// FetchCheapestByName downloads Card Kingdom retail prices from the official CK
// pricelist API and indexes the cheapest listed price per card name.
func FetchCheapestByName(ctx context.Context) (PricelistFetchResult, error) {
	return fetchCheapestFromCKPricelist(ctx)
}
