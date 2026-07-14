package cardkingdom

// considerCheapestListing keeps the lowest USD price per lookup name key.
func considerCheapestListing(cheapest map[string]Listing, listing Listing) {
	if listing.PriceUsd <= 0 {
		return
	}

	for _, nameKey := range ListingNameKeys(listing) {
		existing, ok := cheapest[nameKey]
		if !ok || listing.PriceUsd < existing.PriceUsd {
			cheapest[nameKey] = listing
		}
	}
}
