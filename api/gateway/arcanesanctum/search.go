package arcanesanctum

import (
	"context"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/binderpos"
)

const StoreName = "Arcane Sanctum"
const StoreBaseURL = "https://arcanesanctumtcg.com"
const StoreStorefrontAccessToken = "228ce7e7cffe6623f36634d0ca085e9e"
const StoreShopifyDomain = "30uetm-1y.myshopify.com"
const StoreSearchURL = "/search?q=%s"

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
	return s.BinderposGwy.Search(ctx, 2, s.Name, s.BaseUrl, StoreShopifyDomain, s.SearchUrl, searchStr, StoreStorefrontAccessToken)
}
