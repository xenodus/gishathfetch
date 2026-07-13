package ckprices

import (
	"context"

	"mtg-price-checker-sg/gateway/cardkingdom"
)

const PriceChangeRankingLimit = 20

// PriceChangeListing ranks a Card Kingdom listing by its latest price change.
type PriceChangeListing struct {
	NameKey string `json:"nameKey"`
	cardkingdom.Listing
}

// TopBottomPriceChanges holds the largest price increases and decreases from the
// most recent CK price job run.
type TopBottomPriceChanges struct {
	Top    []PriceChangeListing `json:"top"`
	Bottom []PriceChangeListing `json:"bottom"`
}

// Store persists Card Kingdom cheapest-by-name listings.
type Store interface {
	GetByNameKey(ctx context.Context, nameKey string) (*cardkingdom.Listing, error)
	// GetPriceChangesByPercent returns listings ordered by priceChangePercent.
	// ascending=true is equivalent to SQL ORDER BY priceChangePercent ASC LIMIT n.
	GetPriceChangesByPercent(ctx context.Context, ascending bool, limit int) ([]PriceChangeListing, error)
	// GetPriceChangesByUsd returns listings ordered by priceChangeUsd.
	// ascending=true is equivalent to SQL ORDER BY priceChangeUsd ASC LIMIT n.
	GetPriceChangesByUsd(ctx context.Context, ascending bool, limit int) ([]PriceChangeListing, error)
	GetTopBottomPriceChanges(ctx context.Context) (*TopBottomPriceChanges, error)
	PutAll(ctx context.Context, listings map[string]cardkingdom.Listing) (syncedAt string, err error)
}
