package binderpos

import (
	"mtg-price-checker-sg/pkg/config"
)

const (
	binderposDecklistAPIURL = "https://portal.binderpos.com/external/shopify/decklist"
	binderposDecklistType   = "mtg"
	binderposAttemptTimeout = config.SearchAttemptTimeout
)

type storefrontDecklistRequestItem struct {
	Card     string `json:"card"`
	Quantity int    `json:"quantity"`
}

type storefrontDecklistLine struct {
	// The API can send a boolean validName; binding it would break decode, so it is not mapped.
	Products []storefrontDecklistProduct `json:"products"`
}

type storefrontDecklistProduct struct {
	Title    string                    `json:"title"`
	Name     string                    `json:"name"`
	Handle   string                    `json:"handle"`
	SetName  string                    `json:"setName"`
	Image    string                    `json:"img"`
	Variants []storefrontDecklistStock `json:"variants"`
}

type storefrontDecklistStock struct {
	ShopifyID int64   `json:"shopifyId"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}
