package cardkingdom

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuildCheapestByName(t *testing.T) {
	updatedAt := time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC)
	products := []Product{
		{
			Name:        "Lightning Bolt",
			Edition:     "Alpha",
			PriceRetail: "999.99",
			URL:         "mtg/alpha/lightning-bolt",
			IsFoil:      "false",
		},
		{
			Name:        "Lightning Bolt",
			Edition:     "Fourth Edition",
			PriceRetail: "1.49",
			URL:         "mtg/fourth-edition/lightning-bolt",
			IsFoil:      "false",
		},
		{
			Name:        "Lightning Bolt",
			Edition:     "Modern Masters",
			PriceRetail: "3.99",
			URL:         "mtg/modern-masters/lightning-bolt-foil",
			IsFoil:      "true",
		},
		{
			Name:        "Counterspell",
			Edition:     "Ice Age",
			PriceRetail: "0",
			URL:         "mtg/ice-age/counterspell",
			IsFoil:      "false",
		},
	}

	cheapest := BuildCheapestByName(products, updatedAt)

	require.Len(t, cheapest, 1)
	listing := cheapest["lightning bolt"]
	require.Equal(t, "Lightning Bolt", listing.CardName)
	require.Equal(t, "Fourth Edition", listing.Edition)
	require.InDelta(t, 1.49, listing.PriceUsd, 0.001)
	require.False(t, listing.IsFoil)
	require.Equal(t, "https://www.cardkingdom.com/mtg/fourth-edition/lightning-bolt", listing.URL)
	require.Equal(t, updatedAt.Format(time.RFC3339), listing.UpdatedAt)
}

func TestBuildCheapestByName_DoubleFacedCardIndexesFaceNames(t *testing.T) {
	updatedAt := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	products := []Product{
		{
			Name:        "Jennifer Walters",
			Edition:     "Marvel Super Heroes",
			PriceRetail: "10.99",
			URL:         "mtg/marvel-super-heroes/jennifer-walters",
			IsFoil:      "false",
		},
		{
			Name:        "Jennifer Walters",
			Edition:     "Marvel Super Heroes Variants",
			PriceRetail: "24.99",
			URL:         "mtg/marvel-super-heroes-variants/jennifer-walters-0328-borderless",
			IsFoil:      "false",
		},
	}

	cheapest := BuildCheapestByName(products, updatedAt)

	require.InDelta(t, 10.99, cheapest["jennifer walters"].PriceUsd, 0.001)
	require.Equal(t, "Marvel Super Heroes", cheapest["jennifer walters"].Edition)
}
