package ckprice

import (
	"context"
	"strings"

	"mtg-price-checker-sg/gateway/cardkingdom"
	"mtg-price-checker-sg/gateway/scryfall"
	"mtg-price-checker-sg/store/ckprices"
)

var (
	verifyCardNameFunc = scryfall.VerifyCardName
	fetchCheapestFunc  = cardkingdom.FetchCheapestByName
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

	return store.GetByNameKey(ctx, strings.ToLower(strings.TrimSpace(verifiedName)))
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
