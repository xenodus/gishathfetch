package mtgasia

import (
	"context"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/binderpos"
)

const StoreName = "MTG Asia"
const StoreBaseURL = "https://www.mtg-asia.com"
const StoreShopifyDomain = "mtgasia.myshopify.com"
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
	return s.BinderposGwy.Search(ctx,
		2,
		s.Name,
		s.BaseUrl,
		StoreShopifyDomain,
		s.SearchUrl,
		searchStr,
		false,
	)
}
