package cardkingdom

import "encoding/json"

// Product is a single listing from the Card Kingdom pricelist API.
type Product struct {
	Name        string      `json:"name"`
	Edition     string      `json:"edition"`
	PriceRetail json.Number `json:"price_retail"`
	QtyRetail   json.Number `json:"qty_retail"`
	URL         string      `json:"url"`
	IsFoil      string      `json:"is_foil"`
}

// Listing is the cheapest in-stock Card Kingdom offer for a card name.
type Listing struct {
	CardName           string  `json:"cardName"`
	Edition            string  `json:"edition"`
	PriceUsd           float64 `json:"priceUsd"`
	PriceChangePercent *int    `json:"priceChangePercent,omitempty"`
	URL                string  `json:"url"`
	Quantity           int     `json:"quantity"`
	IsFoil             bool    `json:"isFoil"`
	UpdatedAt          string  `json:"updatedAt,omitempty"`
	SyncedAt           string  `json:"syncedAt,omitempty"`
}
