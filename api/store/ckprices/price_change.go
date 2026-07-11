package ckprices

import (
	"math"

	"mtg-price-checker-sg/gateway/cardkingdom"
)

func computePriceChangePercent(previousPriceUsd, currentPriceUsd float64) *int {
	if previousPriceUsd <= 0 {
		return nil
	}

	changePercent := int(math.Round((currentPriceUsd - previousPriceUsd) / previousPriceUsd * 100))
	return &changePercent
}

func listingsWithPriceChange(
	existing map[string]dynamoRecord,
	listings map[string]cardkingdom.Listing,
) map[string]cardkingdom.Listing {
	enriched := make(map[string]cardkingdom.Listing, len(listings))
	for nameKey, listing := range listings {
		if previous, ok := existing[nameKey]; ok {
			listing.PriceChangePercent = computePriceChangePercent(previous.PriceUsd, listing.PriceUsd)
		}
		enriched[nameKey] = listing
	}
	return enriched
}
