package ckprices

import (
	"math"
	"slices"
	"strings"

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
			previousPriceUsd := previous.PriceUsd
			listing.PreviousPriceUsd = &previousPriceUsd
			listing.PriceChangePercent = computePriceChangePercent(previousPriceUsd, listing.PriceUsd)
		}
		enriched[nameKey] = listing
	}
	return enriched
}

func topBottomPriceChanges(listings []PriceChangeListing, limit int) TopBottomPriceChanges {
	if limit <= 0 {
		limit = PriceChangeRankingLimit
	}

	changed := make([]PriceChangeListing, 0, len(listings))
	for _, listing := range listings {
		if listing.PriceChangePercent == nil {
			continue
		}
		changed = append(changed, listing)
	}

	top := append([]PriceChangeListing(nil), changed...)
	slices.SortFunc(top, func(a, b PriceChangeListing) int {
		if *a.PriceChangePercent != *b.PriceChangePercent {
			return *b.PriceChangePercent - *a.PriceChangePercent
		}
		return strings.Compare(a.NameKey, b.NameKey)
	})
	if len(top) > limit {
		top = top[:limit]
	}

	bottom := append([]PriceChangeListing(nil), changed...)
	slices.SortFunc(bottom, func(a, b PriceChangeListing) int {
		if *a.PriceChangePercent != *b.PriceChangePercent {
			return *a.PriceChangePercent - *b.PriceChangePercent
		}
		return strings.Compare(a.NameKey, b.NameKey)
	})
	if len(bottom) > limit {
		bottom = bottom[:limit]
	}

	return TopBottomPriceChanges{
		Top:    top,
		Bottom: bottom,
	}
}

func priceChangesByPercentFromListings(listings []PriceChangeListing, ascending bool, limit int) []PriceChangeListing {
	rankings := topBottomPriceChanges(listings, limit)
	if ascending {
		return rankings.Bottom
	}
	return rankings.Top
}
