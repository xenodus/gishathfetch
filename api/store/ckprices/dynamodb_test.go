package ckprices

import (
	"testing"
	"time"

	"mtg-price-checker-sg/gateway/cardkingdom"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"
)

func TestDynamoRecordFromListing(t *testing.T) {
	syncedAt := time.Date(2026, 6, 28, 15, 30, 0, 0, time.UTC).Format(time.RFC3339)
	priceChangePercent := 12
	record := dynamoRecordFromListing("lightning bolt", cardkingdom.Listing{
		CardName:           "Lightning Bolt",
		Edition:            "Fourth Edition",
		PriceUsd:           1.49,
		PriceChangePercent: &priceChangePercent,
		URL:                "https://www.cardkingdom.com/mtg/fourth-edition/lightning-bolt",
		Quantity:           0,
		IsFoil:             false,
		UpdatedAt:          "2026-06-28T00:00:00Z",
	}, syncedAt)

	require.Equal(t, "lightning bolt", record.NameKey)
	require.Equal(t, "2026-06-28T00:00:00Z", record.UpdatedAt)
	require.Equal(t, syncedAt, record.SyncedAt)
	require.NotNil(t, record.PriceChangePercent)
	require.Equal(t, 12, *record.PriceChangePercent)
	require.NotNil(t, record.PriceChangeIndexPK)
	require.Equal(t, priceChangeIndexPKValue, *record.PriceChangeIndexPK)
}

func TestDynamoRecordFromListing_OmitsPriceChangeIndexWithoutPercent(t *testing.T) {
	syncedAt := time.Date(2026, 6, 28, 15, 30, 0, 0, time.UTC).Format(time.RFC3339)
	record := dynamoRecordFromListing("new card", cardkingdom.Listing{
		CardName:  "New Card",
		PriceUsd:  1.49,
		UpdatedAt: "2026-06-28T00:00:00Z",
	}, syncedAt)

	require.Nil(t, record.PriceChangePercent)
	require.Nil(t, record.PriceChangeIndexPK)
}

func TestDynamoRecordMarshalIncludesSyncedAt(t *testing.T) {
	syncedAt := time.Date(2026, 6, 28, 15, 30, 0, 0, time.UTC).Format(time.RFC3339)
	record := dynamoRecordFromListing("lightning bolt", cardkingdom.Listing{UpdatedAt: "2026-06-28T00:00:00Z"}, syncedAt)
	item, err := attributevalue.MarshalMap(record)
	require.NoError(t, err)
	require.Contains(t, item, "syncedAt")
	av, ok := item["syncedAt"].(*types.AttributeValueMemberS)
	require.True(t, ok)
	require.Equal(t, syncedAt, av.Value)
}

func TestDynamoRecordMarshalIncludesPriceChangePercent(t *testing.T) {
	syncedAt := time.Date(2026, 6, 28, 15, 30, 0, 0, time.UTC).Format(time.RFC3339)
	priceChangePercent := -8
	record := dynamoRecordFromListing("lightning bolt", cardkingdom.Listing{
		UpdatedAt:          "2026-06-28T00:00:00Z",
		PriceChangePercent: &priceChangePercent,
	}, syncedAt)
	item, err := attributevalue.MarshalMap(record)
	require.NoError(t, err)
	require.Contains(t, item, "priceChangePercent")
	av, ok := item["priceChangePercent"].(*types.AttributeValueMemberN)
	require.True(t, ok)
	require.Equal(t, "-8", av.Value)
}

func TestSyncMetadataRecordMarshal(t *testing.T) {
	syncedAt := time.Date(2026, 6, 28, 15, 30, 0, 0, time.UTC).Format(time.RFC3339)
	item, err := attributevalue.MarshalMap(syncMetadataRecord{
		NameKey:      syncMetadataKey,
		CardName:     syncMetadataLabel,
		SyncedAt:     syncedAt,
		ListingCount: 12345,
	})
	require.NoError(t, err)
	require.Contains(t, item, "syncedAt")
	require.Contains(t, item, "listingCount")
}
