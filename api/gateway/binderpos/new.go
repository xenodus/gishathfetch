package binderpos

import (
	"context"

	"mtg-price-checker-sg/gateway"
)

type Gateway interface {
	// Search runs the BinderPOS strategy chain. storefrontAccessToken enables
	// Storefront GraphQL (dedicated then direct) before HTML scrap when non-empty.
	Search(ctx context.Context, scrapVariant int, storeName, baseUrl, shopifyDomain, searchUrl, searchStr, storefrontAccessToken string) ([]gateway.Card, error)
	Scrap(ctx context.Context, scrapVariant int, storeName, baseUrl, searchUrl, searchStr string) ([]gateway.Card, error)
}

type impl struct{}

func New() Gateway {
	return &impl{}
}
