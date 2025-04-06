package hideout

import (
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/binderpos"
)

const StoreName = "Hideout"
const StoreBaseURL = "https://hideoutcg.com"
const StoreSearchURL = "/search?q=%s"

// const binderposStoreURL = "220022-20.myshopify.com"

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
		3,
		s.Name,
		s.BaseUrl,
		s.SearchUrl,
		searchStr,
	)
}
