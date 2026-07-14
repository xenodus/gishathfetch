package ckprices

import (
	"cmp"
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

func computePriceChangeUsd(previousPriceUsd, currentPriceUsd float64) *float64 {
	if previousPriceUsd <= 0 {
		return nil
	}

	changeUsd := math.Round((currentPriceUsd-previousPriceUsd)*100) / 100
	return &changeUsd
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
			listing.PriceChangeUsd = computePriceChangeUsd(previousPriceUsd, listing.PriceUsd)
		}
		enriched[nameKey] = listing
	}
	return enriched
}

func topBottomPriceChangesByPercent(listings []PriceChangeListing, limit int) TopBottomPriceChanges {
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

func topBottomPriceChangesByUsd(listings []PriceChangeListing, limit int) TopBottomPriceChanges {
	if limit <= 0 {
		limit = PriceChangeRankingLimit
	}

	changed := make([]PriceChangeListing, 0, len(listings))
	for _, listing := range listings {
		if listing.PriceChangeUsd == nil {
			continue
		}
		changed = append(changed, listing)
	}

	top := append([]PriceChangeListing(nil), changed...)
	slices.SortFunc(top, func(a, b PriceChangeListing) int {
		if *a.PriceChangeUsd != *b.PriceChangeUsd {
			return cmp.Compare(*b.PriceChangeUsd, *a.PriceChangeUsd)
		}
		return strings.Compare(a.NameKey, b.NameKey)
	})
	if len(top) > limit {
		top = top[:limit]
	}

	bottom := append([]PriceChangeListing(nil), changed...)
	slices.SortFunc(bottom, func(a, b PriceChangeListing) int {
		if *a.PriceChangeUsd != *b.PriceChangeUsd {
			return cmp.Compare(*a.PriceChangeUsd, *b.PriceChangeUsd)
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

// filterPriceChangesByUsdSign keeps listings whose USD change matches the requested
// direction: increases=true keeps price rises (> 0), increases=false keeps price
// drops (< 0). Listings with a missing or zero change are always excluded.
func filterPriceChangesByUsdSign(listings []PriceChangeListing, increases bool) []PriceChangeListing {
	filtered := make([]PriceChangeListing, 0, len(listings))
	for _, listing := range listings {
		if listing.PriceChangeUsd == nil || *listing.PriceChangeUsd == 0 {
			continue
		}
		if (*listing.PriceChangeUsd > 0) == increases {
			filtered = append(filtered, listing)
		}
	}
	return filtered
}

func priceChangesByPercentFromListings(listings []PriceChangeListing, ascending bool, limit int) []PriceChangeListing {
	rankings := topBottomPriceChangesByPercent(listings, limit)
	if ascending {
		return rankings.Bottom
	}
	return rankings.Top
}

func priceChangesByUsdFromListings(listings []PriceChangeListing, ascending bool, limit int) []PriceChangeListing {
	rankings := topBottomPriceChangesByUsd(listings, limit)
	if ascending {
		return rankings.Bottom
	}
	return rankings.Top
}
