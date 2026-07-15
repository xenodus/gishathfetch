package manapro

import (
	"context"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/binderpos"
)

const StoreName = "Mana Pro"
const StoreBaseURL = "https://sg-manapro.com"
const StoreStorefrontAccessToken = "ba695fbdd730f9fa7b1e8f32e36691cf"
const StoreShopifyDomain = "mana-pro-sg.myshopify.com"
const StoreSearchURL = "/search?type=product&q=%s"

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
