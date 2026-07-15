package gog

import (
	"context"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/binderpos"
)

const StoreName = "Grey Ogre Games"
const StoreBaseURL = "https://www.greyogregames.com"
const StoreStorefrontAccessToken = "80c454d63abad6ad0ebb6b3aaf649fcd"
const StoreShopifyDomain = "grey-ogre-games-singapore.myshopify.com"
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
