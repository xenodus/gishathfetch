package gog

import (
	"encoding/json"
	"log"
	"net/http"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/binderpos"
)

const StoreName = "Grey Ogre Games"
const StoreBaseURL = "https://www.greyogregames.com"
const StoreSearchURL = "/search?q=%s"

const binderposStoreURL = "grey-ogre-games-singapore.myshopify.com"

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
	reqPayload, err := json.Marshal(binderpos.Payload{
		StoreURL:    binderposStoreURL,
		Game:        binderpos.ProductTypeMTG.ToString(),
		Title:       searchStr,
		InstockOnly: true,
	})
	if err != nil {
		return []gateway.Card{}, err
	}

	cards, httpStatusCode, err := s.BinderposGwy.Search(s.Name, s.BaseUrl, reqPayload)
	if err != nil || httpStatusCode != http.StatusOK {
		log.Printf("falling back to scrap for [%s]", s.Name)
		return s.BinderposGwy.Scrap(3, s.Name, s.BaseUrl, s.SearchUrl, searchStr)
	}

	return cards, nil
}
