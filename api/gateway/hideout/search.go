package hideout

import (
	"context"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/binderpos"
)

const StoreName = "Hideout"
const StoreBaseURL = "https://hideoutcg.com"
const StoreStorefrontAccessToken = "986e6f452e7b632be7a14cba965f64a8"
const StoreShopifyDomain = "220022-20.myshopify.com"
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
	return s.BinderposGwy.Search(ctx, 3, s.Name, s.BaseUrl, StoreShopifyDomain, s.SearchUrl, searchStr, StoreStorefrontAccessToken)
}
