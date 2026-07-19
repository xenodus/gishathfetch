package ckprices

import (
	"testing"
	"time"

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

func TestDedupePriceChangeListingsByURL(t *testing.T) {
	usd := func(value float64) *float64 {
		return &value
	}
	sharedURL := "https://www.cardkingdom.com/mtg/example/tony-stark"
	listings := []PriceChangeListing{
		{
			NameKey: "tony stark // the invincible iron man",
			Listing: cardkingdom.Listing{
				CardName:       "Tony Stark // The Invincible Iron Man",
				PriceChangeUsd: usd(26),
				URL:            sharedURL,
			},
		},
		{
			NameKey: "the invincible iron man",
			Listing: cardkingdom.Listing{
				CardName:       "Tony Stark // The Invincible Iron Man",
				PriceChangeUsd: usd(26),
				URL:            sharedURL,
			},
		},
		{
			NameKey: "lightning bolt",
			Listing: cardkingdom.Listing{
				CardName:       "Lightning Bolt",
				PriceChangeUsd: usd(0.50),
				URL:            "https://www.cardkingdom.com/mtg/example/lightning-bolt",
			},
		},
	}

	deduped := dedupePriceChangeListings(listings, 20)
	require.Len(t, deduped, 2)
	require.Equal(t, "tony stark // the invincible iron man", deduped[0].NameKey)
	require.Equal(t, "lightning bolt", deduped[1].NameKey)
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

func TestIsSameUTCDay(t *testing.T) {
	now := time.Date(2026, 7, 19, 15, 30, 0, 0, time.UTC)

	require.True(t, isSameUTCDay("2026-07-19T08:00:00Z", now))
	require.False(t, isSameUTCDay("2026-07-18T23:59:59Z", now))
	require.False(t, isSameUTCDay("not-a-timestamp", now))
}

func TestListingsWithPriceChange(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	existing := map[string]dynamoRecord{
		"lightning bolt": {PriceUsd: 1.00},
		"counterspell":   {PriceUsd: 2.00},
	}

	enriched := listingsWithPriceChange(existing, map[string]cardkingdom.Listing{
		"lightning bolt": {PriceUsd: 1.10},
		"counterspell":   {PriceUsd: 1.80},
		"new card":       {PriceUsd: 5.00},
	}, now)

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

func TestListingsWithPriceChangePreservesSameDayChanges(t *testing.T) {
	now := time.Date(2026, 7, 19, 15, 0, 0, 0, time.UTC)
	previousPrice := 1.00
	changeUsd := 0.10
	changePercent := 10

	existing := map[string]dynamoRecord{
		"lightning bolt": {
			PriceUsd:           1.10,
			PreviousPriceUsd:   &previousPrice,
			PriceChangeUsd:     &changeUsd,
			PriceChangePercent: &changePercent,
			SyncedAt:           "2026-07-19T08:00:00Z",
		},
	}

	enriched := listingsWithPriceChange(existing, map[string]cardkingdom.Listing{
		"lightning bolt": {PriceUsd: 1.15},
	}, now)

	require.Equal(t, 1.15, enriched["lightning bolt"].PriceUsd)
	require.NotNil(t, enriched["lightning bolt"].PreviousPriceUsd)
	require.Equal(t, 1.00, *enriched["lightning bolt"].PreviousPriceUsd)
	require.NotNil(t, enriched["lightning bolt"].PriceChangeUsd)
	require.InDelta(t, 0.10, *enriched["lightning bolt"].PriceChangeUsd, 0.001)
	require.NotNil(t, enriched["lightning bolt"].PriceChangePercent)
	require.Equal(t, 10, *enriched["lightning bolt"].PriceChangePercent)
}
