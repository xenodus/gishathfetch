package mtgasia

import (
	"context"

	"mtg-price-checker-sg/gateway"
)

const StoreName = "MTG Asia"
// StoreBaseURL is the Base44 storefront on mtg-asia.com (card single pages at /card/...).
// Inventory search uses the public Base44 entity API; the legacy Shopify host remains
// available for diagnostics only.
const StoreBaseURL = "https://mtg-asia.com"
const StoreSearchURL = "/search?q=%s"

type Store struct {
	Name      string
	BaseUrl   string
	SearchUrl string
}

func NewLGS() gateway.LGS {
	return Store{
		Name:      StoreName,
		BaseUrl:   StoreBaseURL,
		SearchUrl: StoreSearchURL,
	}
}

func (s Store) Search(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	return searchBase44Inventory(ctx, searchStr)
}
