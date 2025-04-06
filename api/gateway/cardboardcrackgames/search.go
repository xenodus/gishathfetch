package cardboardcrackgames

import (
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/binderpos"
)

const StoreName = "Cardboard Crack Games"
const StoreBaseURL = "https://www.cardboardcrackgames.com"
const StoreSearchURL = "/search?type=product&q=%s"

const binderposStoreURL = "cardboardcrackgames.myshopify.com"

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

func (s Store) Search(searchStr string) ([]gateway.Card, error) {
	return s.BinderposGwy.Scrap(
		2,
		s.Name,
		s.BaseUrl,
		s.SearchUrl,
		searchStr,
	)
}
