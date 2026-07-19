package ckprices

import (
	"cmp"
	"math"
	"slices"
	"strings"
	"time"

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

func isSameUTCDay(syncedAt string, now time.Time) bool {
	syncedTime, err := time.Parse(time.RFC3339, syncedAt)
	if err != nil {
		return false
	}
	syncedUTC := syncedTime.UTC()
	nowUTC := now.UTC()
	return syncedUTC.Year() == nowUTC.Year() &&
		syncedUTC.Month() == nowUTC.Month() &&
		syncedUTC.Day() == nowUTC.Day()
}

func listingsWithPriceChange(
	existing map[string]dynamoRecord,
	listings map[string]cardkingdom.Listing,
	now time.Time,
) map[string]cardkingdom.Listing {
	enriched := make(map[string]cardkingdom.Listing, len(listings))
	for nameKey, listing := range listings {
		if previous, ok := existing[nameKey]; ok {
			if isSameUTCDay(previous.SyncedAt, now) {
				listing.PreviousPriceUsd = previous.PreviousPriceUsd
				listing.PriceChangePercent = previous.PriceChangePercent
				listing.PriceChangeUsd = previous.PriceChangeUsd
			} else {
				previousPriceUsd := previous.PriceUsd
				listing.PreviousPriceUsd = &previousPriceUsd
				listing.PriceChangePercent = computePriceChangePercent(previousPriceUsd, listing.PriceUsd)
				listing.PriceChangeUsd = computePriceChangeUsd(previousPriceUsd, listing.PriceUsd)
			}
		}
		enriched[nameKey] = listing
	}
	return enriched
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
	top = dedupePriceChangeListings(top, limit)

	bottom := append([]PriceChangeListing(nil), changed...)
	slices.SortFunc(bottom, func(a, b PriceChangeListing) int {
		if *a.PriceChangeUsd != *b.PriceChangeUsd {
			return cmp.Compare(*a.PriceChangeUsd, *b.PriceChangeUsd)
		}
		return strings.Compare(a.NameKey, b.NameKey)
	})
	bottom = dedupePriceChangeListings(bottom, limit)

	return TopBottomPriceChanges{
		Top:    top,
		Bottom: bottom,
	}
}

func priceChangeListingDedupeKey(listing PriceChangeListing) string {
	if url := strings.TrimSpace(listing.URL); url != "" {
		return "url:" + url
	}

	cardName := cardkingdom.NormalizeNameKey(listing.CardName)
	edition := strings.ToLower(strings.TrimSpace(listing.Edition))
	if cardName != "" {
		foil := "nonfoil"
		if listing.IsFoil {
			foil = "foil"
		}
		return "listing:" + cardName + "|" + edition + "|" + foil
	}

	if listing.NameKey != "" {
		return "nameKey:" + listing.NameKey
	}

	return ""
}

func dedupePriceChangeListings(listings []PriceChangeListing, limit int) []PriceChangeListing {
	if limit <= 0 {
		limit = PriceChangeRankingLimit
	}

	seen := make(map[string]struct{}, len(listings))
	deduped := make([]PriceChangeListing, 0, limit)
	for _, listing := range listings {
		key := priceChangeListingDedupeKey(listing)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, listing)
		if len(deduped) >= limit {
			break
		}
	}
	return deduped
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

func priceChangesByUsdFromListings(listings []PriceChangeListing, ascending bool, limit int) []PriceChangeListing {
	rankings := topBottomPriceChangesByUsd(listings, limit)
	if ascending {
		return rankings.Bottom
	}
	return rankings.Top
}
