package cardkingdom

import (
	"strings"
	"time"
)

const listingBaseURL = "https://www.cardkingdom.com/"

// BuildCheapestByName returns the lowest retail USD price per card name.
func BuildCheapestByName(products []Product, updatedAt time.Time) map[string]Listing {
	cheapestByName := make(map[string]Listing)
	updatedAtValue := updatedAt.UTC().Format(time.RFC3339)

	for _, product := range products {
		nameKey := strings.TrimSpace(strings.ToLower(product.Name))
		if nameKey == "" {
			continue
		}

		priceUsd, err := product.PriceRetail.Float64()
		if err != nil || priceUsd <= 0 {
			continue
		}

		quantity64, _ := product.QtyRetail.Int64()
		listing := Listing{
			CardName:  product.Name,
			Edition:   product.Edition,
			PriceUsd:  priceUsd,
			URL:       listingBaseURL + strings.TrimPrefix(product.URL, "/"),
			Quantity:  int(quantity64),
			IsFoil:    strings.EqualFold(strings.TrimSpace(product.IsFoil), "true"),
			UpdatedAt: updatedAtValue,
		}

		existing, ok := cheapestByName[nameKey]
		if !ok || priceUsd < existing.PriceUsd {
			cheapestByName[nameKey] = listing
		}
	}

	return cheapestByName
}
