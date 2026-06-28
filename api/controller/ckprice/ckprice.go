package ckprice

import (
	"context"
	"log"
	"strings"
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
func GetLatestPrice(ctx context.Context, store ckprices.Store, query string) (*cardkingdom.Listing, error) {
	verifiedName, err := verifyCardNameFunc(ctx, query)
	if err != nil {
		return nil, err
	}
	if verifiedName == "" {
		return nil, nil
	}

	listing, err := store.GetByNameKey(ctx, strings.ToLower(strings.TrimSpace(verifiedName)))
	if err != nil || listing == nil {
		return nil, err
	}
	if !listingIsFresh(listing, nowFunc()) {
		return nil, nil
	}
	return listing, nil
}

func listingIsFresh(listing *cardkingdom.Listing, now time.Time) bool {
	if listing == nil || listing.UpdatedAt == "" {
		return false
	}
	updatedAt, err := time.Parse(time.RFC3339, listing.UpdatedAt)
	if err != nil {
		return false
	}
	return !now.After(updatedAt.Add(config.CKPriceMaxAge))
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
