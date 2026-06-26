package arcanesanctum

import (
	"context"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/binderpos"
)

const StoreName = binderpos.ArcaneSanctumStoreName
const StoreBaseURL = "https://arcanesanctumtcg.com"
const StoreShopifyDomain = "30uetm-1y.myshopify.com"
const StoreSearchURL = "/search?q=%s"

// ScrapOnly skips the shared BinderPOS decklist portal while retaining the
// Shopify domain mapping for documentation and live integration tests.
const ScrapOnly = true

type Store struct {
	Name         string
	BaseUrl      string
	SearchUrl    string
	BinderposGwy binderpos.Gateway
}

func NewLGS() gateway.LGS {
	return Store{
		Name:         StoreName,
		BaseUrl:      StoreBaseURL,
		SearchUrl:    StoreSearchURL,
		BinderposGwy: binderpos.New(),
	}
}

func (s Store) Search(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	return s.BinderposGwy.Search(ctx,
		5,
		s.Name,
		s.BaseUrl,
		StoreShopifyDomain,
		s.SearchUrl,
		searchStr,
		ScrapOnly,
	)
}
