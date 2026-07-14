package ckprices

import (
	"strconv"
	"testing"

	"mtg-price-checker-sg/gateway/cardkingdom"

	"github.com/stretchr/testify/require"
)

func TestComputePriceChangePercent(t *testing.T) {
	t.Run("increase", func(t *testing.T) {
		change := computePriceChangePercent(10, 12)
		require.NotNil(t, change)
		require.Equal(t, 20, *change)
	})

	t.Run("decrease", func(t *testing.T) {
		change := computePriceChangePercent(10, 8)
		require.NotNil(t, change)
		require.Equal(t, -20, *change)
	})

	t.Run("rounds to nearest integer", func(t *testing.T) {
		change := computePriceChangePercent(3, 3.34)
		require.NotNil(t, change)
		require.Equal(t, 11, *change)
	})

	t.Run("no change", func(t *testing.T) {
		change := computePriceChangePercent(4.99, 4.99)
		require.NotNil(t, change)
		require.Equal(t, 0, *change)
	})

	t.Run("missing previous price", func(t *testing.T) {
		require.Nil(t, computePriceChangePercent(0, 4.99))
	})
}

func TestComputePriceChangeUsd(t *testing.T) {
	t.Run("increase", func(t *testing.T) {
		change := computePriceChangeUsd(10, 12)
		require.NotNil(t, change)
		require.InDelta(t, 2, *change, 0.001)
	})

	t.Run("decrease", func(t *testing.T) {
		change := computePriceChangeUsd(10, 8)
		require.NotNil(t, change)
		require.InDelta(t, -2, *change, 0.001)
	})

	t.Run("rounds to cents", func(t *testing.T) {
		change := computePriceChangeUsd(1.00, 1.10)
		require.NotNil(t, change)
		require.InDelta(t, 0.10, *change, 0.001)
	})

	t.Run("no change", func(t *testing.T) {
		change := computePriceChangeUsd(4.99, 4.99)
		require.NotNil(t, change)
		require.InDelta(t, 0, *change, 0.001)
	})

	t.Run("missing previous price", func(t *testing.T) {
		require.Nil(t, computePriceChangeUsd(0, 4.99))
	})
}

func TestPriceChangesByPercentFromListings(t *testing.T) {
	percent := func(value int) *int {
		return &value
	}
	listing := func(nameKey string, change int) PriceChangeListing {
		return PriceChangeListing{
			NameKey: nameKey,
			Listing: cardkingdom.Listing{
				CardName:           nameKey,
				PriceChangePercent: percent(change),
			},
		}
	}

	listings := []PriceChangeListing{
		listing("a", 30),
		listing("b", 10),
		listing("c", -30),
		listing("d", -10),
	}

	bottom := priceChangesByPercentFromListings(listings, true, 2)
	require.Len(t, bottom, 2)
	require.Equal(t, -30, *bottom[0].PriceChangePercent)
	require.Equal(t, -10, *bottom[1].PriceChangePercent)

	top := priceChangesByPercentFromListings(listings, false, 2)
	require.Len(t, top, 2)
	require.Equal(t, 30, *top[0].PriceChangePercent)
	require.Equal(t, 10, *top[1].PriceChangePercent)
}

func TestTopBottomPriceChanges(t *testing.T) {
	percent := func(value int) *int {
		return &value
	}
	listing := func(nameKey string, change int) PriceChangeListing {
		return PriceChangeListing{
			NameKey: nameKey,
			Listing: cardkingdom.Listing{
				CardName:           nameKey,
				PriceChangePercent: percent(change),
			},
		}
	}

	listings := make([]PriceChangeListing, 0, 51)
	for i := 1; i <= 25; i++ {
		listings = append(listings, listing("increase-"+strconv.Itoa(i), i))
	}
	for i := 1; i <= 25; i++ {
		listings = append(listings, listing("decrease-"+strconv.Itoa(i), -i))
	}
	listings = append(listings, PriceChangeListing{
		NameKey: "no-change",
		Listing: cardkingdom.Listing{CardName: "No Change"},
	})

	rankings := topBottomPriceChangesByPercent(listings, PriceChangeRankingLimit)

	require.Len(t, rankings.Top, PriceChangeRankingLimit)
	require.Equal(t, 25, *rankings.Top[0].PriceChangePercent)
	require.Equal(t, 6, *rankings.Top[19].PriceChangePercent)

	require.Len(t, rankings.Bottom, PriceChangeRankingLimit)
	require.Equal(t, -25, *rankings.Bottom[0].PriceChangePercent)
	require.Equal(t, -6, *rankings.Bottom[19].PriceChangePercent)
}

func TestPriceChangesByUsdFromListings(t *testing.T) {
	usd := func(value float64) *float64 {
		return &value
	}
	listing := func(nameKey string, change float64) PriceChangeListing {
		return PriceChangeListing{
			NameKey: nameKey,
			Listing: cardkingdom.Listing{
				CardName:       nameKey,
				PriceChangeUsd: usd(change),
			},
		}
	}

	listings := []PriceChangeListing{
		listing("a", 30),
		listing("b", 10),
		listing("c", -30),
		listing("d", -10),
	}

	bottom := priceChangesByUsdFromListings(listings, true, 2)
	require.Len(t, bottom, 2)
	require.InDelta(t, -30, *bottom[0].PriceChangeUsd, 0.001)
	require.InDelta(t, -10, *bottom[1].PriceChangeUsd, 0.001)

	top := priceChangesByUsdFromListings(listings, false, 2)
	require.Len(t, top, 2)
	require.InDelta(t, 30, *top[0].PriceChangeUsd, 0.001)
	require.InDelta(t, 10, *top[1].PriceChangeUsd, 0.001)
}

func TestTopBottomPriceChangesByUsd(t *testing.T) {
	usd := func(value float64) *float64 {
		return &value
	}
	percent := func(value int) *int {
		return &value
	}
	listing := func(nameKey string, changeUsd float64, changePercent int) PriceChangeListing {
		return PriceChangeListing{
			NameKey: nameKey,
			Listing: cardkingdom.Listing{
				CardName:           nameKey,
				PriceChangeUsd:     usd(changeUsd),
				PriceChangePercent: percent(changePercent),
			},
		}
	}

	// USD ranking should prefer the larger absolute dollar move even when percent is lower.
	listings := []PriceChangeListing{
		listing("small-base-big-percent", 0.50, 50),
		listing("big-base-small-percent", 10.00, 10),
		listing("small-drop", -0.25, -25),
		listing("big-drop", -8.00, -8),
	}

	rankings := topBottomPriceChangesByUsd(listings, 1)
	require.Len(t, rankings.Top, 1)
	require.Equal(t, "big-base-small-percent", rankings.Top[0].NameKey)
	require.InDelta(t, 10.00, *rankings.Top[0].PriceChangeUsd, 0.001)

	require.Len(t, rankings.Bottom, 1)
	require.Equal(t, "big-drop", rankings.Bottom[0].NameKey)
	require.InDelta(t, -8.00, *rankings.Bottom[0].PriceChangeUsd, 0.001)
}

func TestFilterPriceChangesByUsdSign(t *testing.T) {
	usd := func(value float64) *float64 {
		return &value
	}
	listings := []PriceChangeListing{
		{NameKey: "riser", Listing: cardkingdom.Listing{PriceChangeUsd: usd(2.50)}},
		{NameKey: "drop", Listing: cardkingdom.Listing{PriceChangeUsd: usd(-1.25)}},
		{NameKey: "flat", Listing: cardkingdom.Listing{PriceChangeUsd: usd(0)}},
		{NameKey: "missing", Listing: cardkingdom.Listing{}},
	}

	increases := filterPriceChangesByUsdSign(listings, true)
	require.Len(t, increases, 1)
	require.Equal(t, "riser", increases[0].NameKey)

	drops := filterPriceChangesByUsdSign(listings, false)
	require.Len(t, drops, 1)
	require.Equal(t, "drop", drops[0].NameKey)
}

func TestPriceChangeListingFromRecord(t *testing.T) {
	change := 15
	listing, ok := priceChangeListingFromRecord(dynamoRecord{
		NameKey:            "lightning bolt",
		CardName:           "Lightning Bolt",
		PriceUsd:           1.49,
		PriceChangePercent: &change,
	})
	require.True(t, ok)
	require.Equal(t, "lightning bolt", listing.NameKey)
	require.Equal(t, 15, *listing.PriceChangePercent)

	_, ok = priceChangeListingFromRecord(dynamoRecord{NameKey: syncMetadataKey})
	require.False(t, ok)

	_, ok = priceChangeListingFromRecord(dynamoRecord{NameKey: "new card", CardName: "New Card"})
	require.False(t, ok)
}

func TestPriceChangeListingFromRecordByUsd(t *testing.T) {
	change := 0.15
	listing, ok := priceChangeListingFromRecordByUsd(dynamoRecord{
		NameKey:        "lightning bolt",
		CardName:       "Lightning Bolt",
		PriceUsd:       1.49,
		PriceChangeUsd: &change,
	})
	require.True(t, ok)
	require.Equal(t, "lightning bolt", listing.NameKey)
	require.InDelta(t, 0.15, *listing.PriceChangeUsd, 0.001)

	_, ok = priceChangeListingFromRecordByUsd(dynamoRecord{NameKey: syncMetadataKey})
	require.False(t, ok)

	_, ok = priceChangeListingFromRecordByUsd(dynamoRecord{NameKey: "new card", CardName: "New Card"})
	require.False(t, ok)
}

func TestListingsWithPriceChange(t *testing.T) {
	existing := map[string]dynamoRecord{
		"lightning bolt": {PriceUsd: 1.00},
		"counterspell":   {PriceUsd: 2.00},
	}

	enriched := listingsWithPriceChange(existing, map[string]cardkingdom.Listing{
		"lightning bolt": {PriceUsd: 1.10},
		"counterspell":   {PriceUsd: 1.80},
		"new card":       {PriceUsd: 5.00},
	})

	require.NotNil(t, enriched["lightning bolt"].PriceChangePercent)
	require.Equal(t, 10, *enriched["lightning bolt"].PriceChangePercent)
	require.NotNil(t, enriched["lightning bolt"].PriceChangeUsd)
	require.InDelta(t, 0.10, *enriched["lightning bolt"].PriceChangeUsd, 0.001)
	require.NotNil(t, enriched["lightning bolt"].PreviousPriceUsd)
	require.Equal(t, 1.00, *enriched["lightning bolt"].PreviousPriceUsd)
	require.NotNil(t, enriched["counterspell"].PriceChangePercent)
	require.Equal(t, -10, *enriched["counterspell"].PriceChangePercent)
	require.NotNil(t, enriched["counterspell"].PriceChangeUsd)
	require.InDelta(t, -0.20, *enriched["counterspell"].PriceChangeUsd, 0.001)
	require.NotNil(t, enriched["counterspell"].PreviousPriceUsd)
	require.Equal(t, 2.00, *enriched["counterspell"].PreviousPriceUsd)
	require.Nil(t, enriched["new card"].PriceChangePercent)
	require.Nil(t, enriched["new card"].PriceChangeUsd)
	require.Nil(t, enriched["new card"].PreviousPriceUsd)
}
