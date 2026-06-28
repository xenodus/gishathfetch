package ckprice

import (
	"context"
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

// RefreshPrices downloads the Card Kingdom pricelist and upserts the DynamoDB index.
func RefreshPrices(ctx context.Context, store ckprices.Store) (int, error) {
	listings, err := fetchCheapestFunc(ctx)
	if err != nil {
		return 0, err
	}
	if err := store.PutAll(ctx, listings); err != nil {
		return 0, err
	}
	return len(listings), nil
}
