package ckprices

import (
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
	require.NotNil(t, enriched["counterspell"].PriceChangePercent)
	require.Equal(t, -10, *enriched["counterspell"].PriceChangePercent)
	require.Nil(t, enriched["new card"].PriceChangePercent)
}
