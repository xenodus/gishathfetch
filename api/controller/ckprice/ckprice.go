package ckprice

import (
	"context"
	"log"
	"time"

	"mtg-price-checker-sg/gateway/cardkingdom"
	"mtg-price-checker-sg/gateway/scryfall"
	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/store/ckprices"
)

var (
	verifyCardNameFunc = scryfall.VerifyCardName
	fetchCheapestFunc  = cardkingdom.FetchCheapestByName
	nowFunc            = time.Now
)

// GetLatestPrice verifies the query against Scryfall and returns the cheapest CK listing.
// For double-faced cards it always checks the combined name, front face, and back
// face together and returns the lowest fresh price across all three.
func GetLatestPrice(ctx context.Context, store ckprices.Store, query string) (*cardkingdom.Listing, error) {
	verifiedName, err := verifyCardNameFunc(ctx, query)
	if err != nil {
		return nil, err
	}
	if verifiedName == "" {
		return nil, nil
	}

	now := nowFunc()
	return cheapestFreshListing(ctx, store, cardkingdom.PriceLookupKeys(verifiedName), now)
}

func cheapestFreshListing(
	ctx context.Context,
	store ckprices.Store,
	nameKeys []string,
	now time.Time,
) (*cardkingdom.Listing, error) {
	var best *cardkingdom.Listing
	for _, nameKey := range nameKeys {
		listing, err := store.GetByNameKey(ctx, nameKey)
		if err != nil {
			return nil, err
		}
		if listing == nil || !listingIsFresh(listing, now) {
			continue
		}
		if best == nil || listing.PriceUsd < best.PriceUsd {
			best = listing
		}
	}
	return best, nil
}

func listingIsFresh(listing *cardkingdom.Listing, now time.Time) bool {
	if listing == nil {
		return false
	}

	// Prefer DynamoDB sync time over MTGJSON's calendar price date.
	freshnessSource := listing.SyncedAt
	if freshnessSource == "" {
		freshnessSource = listing.UpdatedAt
	}
	if freshnessSource == "" {
		return false
	}

	freshnessTime, err := time.Parse(time.RFC3339, freshnessSource)
	if err != nil {
		return false
	}
	return !now.After(freshnessTime.Add(config.CKPriceMaxAge))
}

// RefreshPrices downloads Card Kingdom retail prices from MTGJSON and upserts the DynamoDB index.
func RefreshPrices(ctx context.Context, store ckprices.Store) (int, error) {
	log.Printf("ck price refresh: fetching mtgjson prices")
	fetchStarted := time.Now()

	listings, err := fetchCheapestFunc(ctx)
	if err != nil {
		return 0, err
	}

	log.Printf("ck price refresh: fetched mtgjson prices listings=%d duration=%s", len(listings), time.Since(fetchStarted).Round(time.Millisecond))

	log.Printf("ck price refresh: writing dynamodb listings=%d", len(listings))
	writeStarted := time.Now()
	syncedAt, err := store.PutAll(ctx, listings)
	if err != nil {
		return 0, err
	}
	log.Printf("ck price refresh: wrote dynamodb listings=%d syncedAt=%s duration=%s", len(listings), syncedAt, time.Since(writeStarted).Round(time.Millisecond))

	return len(listings), nil
}
