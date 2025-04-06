package tefuda

import (
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/binderpos"
)

const StoreName = "Tefuda"
const StoreBaseURL = "https://tefudagames.com"
const StoreSearchURL = "/search?q=%s"

// const binderposStoreURL = "bacc1b-3.myshopify.com"

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
		4,
		s.Name,
		s.BaseUrl,
		s.SearchUrl,
		searchStr,
	)
}
