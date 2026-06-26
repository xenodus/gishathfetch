package ckprices

import (
	"context"

	"mtg-price-checker-sg/gateway/cardkingdom"
)

// Store persists Card Kingdom cheapest-by-name listings.
type Store interface {
	GetByNameKey(ctx context.Context, nameKey string) (*cardkingdom.Listing, error)
	PutAll(ctx context.Context, listings map[string]cardkingdom.Listing) error
}
