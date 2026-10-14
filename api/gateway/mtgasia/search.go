package mtgasia

import (
	"context"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/binderpos"
)

const StoreName = "MTG Asia"
// StoreBaseURL is the Shopify storefront used for product deep links. The public
// site (https://mtg-asia.com) is a Base44 app and no longer serves BinderPOS HTML.
const StoreBaseURL = "https://new.mtg-asia.com"
const StoreStorefrontAccessToken = "199d9ab7d26aaba337bd43110a51265f"
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
	return s.BinderposGwy.Search(ctx, 2, s.Name, s.BaseUrl, StoreShopifyDomain, s.SearchUrl, searchStr, StoreStorefrontAccessToken)
}
