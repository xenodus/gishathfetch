package cardaffinity

import (
	"context"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/binderpos"
)

const StoreName = "Card Affinity"
const StoreBaseURL = "https://card-affinity.com"
const StoreShopifyDomain = "563304-2.myshopify.com"
const StoreSearchURL = "/search?q=%s"
// StoreStorefrontAccessToken is published in the shop theme for Storefront API use.
const StoreStorefrontAccessToken = "28c94c0dd51e167fe6bc5b7f4151e099"

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
