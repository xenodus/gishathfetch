package gog

import (
	"context"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/shopifysuggest"
)

const StoreName = "Grey Ogre Games"
const StoreBaseURL = "https://www.greyogregames.com"

type Store struct {
	Name    string
	BaseUrl string
}

func NewLGS() gateway.LGS {
	return Store{
		Name:    StoreName,
		BaseUrl: StoreBaseURL,
	}
}

func (s Store) Search(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	return shopifysuggest.Search(ctx, shopifysuggest.Options{
		Config: shopifysuggest.Config{
			StoreName: s.Name,
			BaseURL:   s.BaseUrl,
		},
		SearchStr:       searchStr,
		BuildQuery:      shopifysuggest.PlainQuery,
		QueryValues:     shopifysuggest.BinderposQueryValues,
		MapProduct:      shopifysuggest.MapBinderposSetExtraProduct,
		ResolveVariants: true,
	})
}
