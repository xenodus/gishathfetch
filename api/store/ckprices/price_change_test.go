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

	rankings := topBottomPriceChanges(listings, PriceChangeRankingLimit)

	require.Len(t, rankings.Top, PriceChangeRankingLimit)
	require.Equal(t, 25, *rankings.Top[0].PriceChangePercent)
	require.Equal(t, 6, *rankings.Top[19].PriceChangePercent)

	require.Len(t, rankings.Bottom, PriceChangeRankingLimit)
	require.Equal(t, -25, *rankings.Bottom[0].PriceChangePercent)
	require.Equal(t, -6, *rankings.Bottom[19].PriceChangePercent)
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
	require.NotNil(t, enriched["lightning bolt"].PreviousPriceUsd)
	require.Equal(t, 1.00, *enriched["lightning bolt"].PreviousPriceUsd)
	require.NotNil(t, enriched["counterspell"].PriceChangePercent)
	require.Equal(t, -10, *enriched["counterspell"].PriceChangePercent)
	require.NotNil(t, enriched["counterspell"].PreviousPriceUsd)
	require.Equal(t, 2.00, *enriched["counterspell"].PreviousPriceUsd)
	require.Nil(t, enriched["new card"].PriceChangePercent)
	require.Nil(t, enriched["new card"].PreviousPriceUsd)
}
