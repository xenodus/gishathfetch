package ckprices

import (
	"testing"
	"time"

	"mtg-price-checker-sg/gateway/cardkingdom"

	"github.com/stretchr/testify/require"
)

func TestDynamoRecordFromListing(t *testing.T) {
	syncedAt := time.Date(2026, 6, 28, 15, 30, 0, 0, time.UTC).Format(time.RFC3339)
	record := dynamoRecordFromListing("lightning bolt", cardkingdom.Listing{
		CardName:  "Lightning Bolt",
		Edition:   "Fourth Edition",
		PriceUsd:  1.49,
		URL:       "https://www.cardkingdom.com/mtg/fourth-edition/lightning-bolt",
		Quantity:  0,
		IsFoil:    false,
		UpdatedAt: "2026-06-28T00:00:00Z",
	}, syncedAt)

	require.Equal(t, "lightning bolt", record.NameKey)
	require.Equal(t, "2026-06-28T00:00:00Z", record.UpdatedAt)
	require.Equal(t, syncedAt, record.SyncedAt)
}
