package binderpos

import (
	"net/http"
	"strings"

	"mtg-price-checker-sg/gateway"
)

type ProductType string

const ProductTypeMTG ProductType = "mtg"

func (p ProductType) ToString() string {
	return string(p)
}

type Gateway interface {
	Search(storeName, storeBaseURL string, payload []byte) ([]gateway.Card, int, error)
	Scrap(scrapVariant int, storeName, baseUrl, searchUrl, searchStr string) ([]gateway.Card, error)
}

type impl struct {
	searchClient http.Client
	apiEndpoint  string
}

func New() Gateway {
	return &impl{
		searchClient: http.Client{},
		apiEndpoint:  apiEndpoint,
	}
}

func NewWithApiUrl(apiUrl string) Gateway {
	return &impl{
		searchClient: http.Client{},
		apiEndpoint:  strings.TrimSpace(apiUrl),
	}
}
