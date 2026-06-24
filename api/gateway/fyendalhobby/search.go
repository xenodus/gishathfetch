package fyendalhobby

import (
	"context"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/shopifysuggest"
)

const StoreName = "Fyendal Hobby"
const StoreBaseURL = "https://fyendalhobby.com"

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
		SearchStr:   searchStr,
		BuildQuery:  shopifysuggest.FyendalQuery,
		QueryValues: shopifysuggest.FyendalQueryValues,
		MapProduct:  shopifysuggest.MapFyendalProduct,
	})
}
