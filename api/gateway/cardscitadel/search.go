package cardscitadel

import (
	"context"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/binderpos"
)

const StoreName = "Cards Citadel"
const StoreBaseURL = "https://cardscitadel.com"
const StoreStorefrontAccessToken = "b68bd33b7d819fc110eb25a07988cc8e"
const StoreShopifyDomain = "card-citadel.myshopify.com"
const StoreSearchURL = "/search?q=*%s*"

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
	return s.BinderposGwy.Search(ctx, 1, s.Name, s.BaseUrl, StoreShopifyDomain, s.SearchUrl, searchStr, StoreStorefrontAccessToken)
}
