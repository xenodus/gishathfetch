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
			QtyRetail:   "1",
			URL:         "mtg/alpha/lightning-bolt",
			IsFoil:      "false",
		},
		{
			Name:        "Lightning Bolt",
			Edition:     "Fourth Edition",
			PriceRetail: "1.49",
			QtyRetail:   "12",
			URL:         "mtg/fourth-edition/lightning-bolt",
			IsFoil:      "false",
		},
		{
			Name:        "Lightning Bolt",
			Edition:     "Modern Masters",
			PriceRetail: "3.99",
			QtyRetail:   "4",
			URL:         "mtg/modern-masters/lightning-bolt-foil",
			IsFoil:      "true",
		},
		{
			Name:        "Counterspell",
			Edition:     "Ice Age",
			PriceRetail: "0",
			QtyRetail:   "1",
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
	require.Equal(t, 12, listing.Quantity)
	require.False(t, listing.IsFoil)
	require.Equal(t, "https://www.cardkingdom.com/mtg/fourth-edition/lightning-bolt", listing.URL)
	require.Equal(t, updatedAt.Format(time.RFC3339), listing.UpdatedAt)
}
