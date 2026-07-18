package cardkingdom

// Listing is the cheapest Card Kingdom retail offer for a card name.
type Listing struct {
	CardName           string   `json:"cardName"`
	Edition            string   `json:"edition"`
	PriceUsd           float64  `json:"priceUsd"`
	PreviousPriceUsd   *float64 `json:"previousPriceUsd,omitempty"`
	PriceChangePercent *int     `json:"priceChangePercent,omitempty"`
	PriceChangeUsd     *float64 `json:"priceChangeUsd,omitempty"`
	URL                string   `json:"url"`
	IsFoil             bool     `json:"isFoil"`
	InStock            *bool    `json:"inStock,omitempty"`
	UpdatedAt          string   `json:"updatedAt,omitempty"`
	SyncedAt           string   `json:"syncedAt,omitempty"`
}
