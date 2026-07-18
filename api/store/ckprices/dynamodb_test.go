package ckprices

import (
	"context"
	"errors"
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
	priceChangeUsd := 0.16
	previousPriceUsd := 1.33
	record := dynamoRecordFromListing("lightning bolt", cardkingdom.Listing{
		CardName:           "Lightning Bolt",
		Edition:            "Fourth Edition",
		PriceUsd:           1.49,
		PreviousPriceUsd:   &previousPriceUsd,
		PriceChangePercent: &priceChangePercent,
		PriceChangeUsd:     &priceChangeUsd,
		URL:                "https://www.cardkingdom.com/mtg/fourth-edition/lightning-bolt",
		IsFoil:             false,
		UpdatedAt:          "2026-06-28T00:00:00Z",
	}, syncedAt)

	require.Equal(t, "lightning bolt", record.NameKey)
	require.Equal(t, "2026-06-28T00:00:00Z", record.UpdatedAt)
	require.Equal(t, syncedAt, record.SyncedAt)
	require.NotNil(t, record.PreviousPriceUsd)
	require.Equal(t, 1.33, *record.PreviousPriceUsd)
	require.NotNil(t, record.PriceChangePercent)
	require.Equal(t, 12, *record.PriceChangePercent)
	require.NotNil(t, record.PriceChangeUsd)
	require.InDelta(t, 0.16, *record.PriceChangeUsd, 0.001)
	require.NotNil(t, record.PriceChangeIndexPK)
	require.Equal(t, priceChangeIndexPKValue, *record.PriceChangeIndexPK)
}

func TestDynamoRecordFromListing_OmitsPriceChangeIndexWithoutUsd(t *testing.T) {
	syncedAt := time.Date(2026, 6, 28, 15, 30, 0, 0, time.UTC).Format(time.RFC3339)
	priceChangePercent := 12
	record := dynamoRecordFromListing("new card", cardkingdom.Listing{
		CardName:           "New Card",
		PriceUsd:           1.49,
		PriceChangePercent: &priceChangePercent,
		UpdatedAt:          "2026-06-28T00:00:00Z",
	}, syncedAt)

	require.NotNil(t, record.PriceChangePercent)
	require.Nil(t, record.PriceChangeUsd)
	require.Nil(t, record.PriceChangeIndexPK)
}

func TestDynamoRecordMarshalIncludesInStock(t *testing.T) {
	syncedAt := time.Date(2026, 6, 28, 15, 30, 0, 0, time.UTC).Format(time.RFC3339)
	inStock := true
	record := dynamoRecordFromListing("lightning bolt", cardkingdom.Listing{
		UpdatedAt: "2026-06-28T00:00:00Z",
		InStock:   &inStock,
	}, syncedAt)
	item, err := attributevalue.MarshalMap(record)
	require.NoError(t, err)
	require.Contains(t, item, "inStock")
	av, ok := item["inStock"].(*types.AttributeValueMemberBOOL)
	require.True(t, ok)
	require.True(t, av.Value)
}

func TestDynamoRecordMarshalOmitsQuantity(t *testing.T) {
	syncedAt := time.Date(2026, 6, 28, 15, 30, 0, 0, time.UTC).Format(time.RFC3339)
	record := dynamoRecordFromListing("lightning bolt", cardkingdom.Listing{UpdatedAt: "2026-06-28T00:00:00Z"}, syncedAt)
	item, err := attributevalue.MarshalMap(record)
	require.NoError(t, err)
	require.NotContains(t, item, "quantity")
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

func TestDynamoRecordMarshalIncludesPriceChangeUsd(t *testing.T) {
	syncedAt := time.Date(2026, 6, 28, 15, 30, 0, 0, time.UTC).Format(time.RFC3339)
	priceChangeUsd := 0.25
	previousPriceUsd := 1.00
	record := dynamoRecordFromListing("lightning bolt", cardkingdom.Listing{
		UpdatedAt:        "2026-06-28T00:00:00Z",
		PreviousPriceUsd: &previousPriceUsd,
		PriceChangeUsd:   &priceChangeUsd,
	}, syncedAt)
	item, err := attributevalue.MarshalMap(record)
	require.NoError(t, err)
	require.Contains(t, item, "priceChangeUsd")
	av, ok := item["priceChangeUsd"].(*types.AttributeValueMemberN)
	require.True(t, ok)
	require.Equal(t, "0.25", av.Value)
}

func TestDynamoRecordMarshalIncludesPriceChangePercent(t *testing.T) {
	syncedAt := time.Date(2026, 6, 28, 15, 30, 0, 0, time.UTC).Format(time.RFC3339)
	priceChangePercent := -8
	previousPriceUsd := 1.62
	record := dynamoRecordFromListing("lightning bolt", cardkingdom.Listing{
		UpdatedAt:          "2026-06-28T00:00:00Z",
		PreviousPriceUsd:   &previousPriceUsd,
		PriceChangePercent: &priceChangePercent,
	}, syncedAt)
	item, err := attributevalue.MarshalMap(record)
	require.NoError(t, err)
	require.Contains(t, item, "priceChangePercent")
	av, ok := item["priceChangePercent"].(*types.AttributeValueMemberN)
	require.True(t, ok)
	require.Equal(t, "-8", av.Value)
	require.Contains(t, item, "previousPriceUsd")
	previousAV, ok := item["previousPriceUsd"].(*types.AttributeValueMemberN)
	require.True(t, ok)
	require.Equal(t, "1.62", previousAV.Value)
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

func TestIsDynamoDBThrottleError(t *testing.T) {
	require.True(t, isDynamoDBThrottleError(&types.ThrottlingException{}))
	require.True(t, isDynamoDBThrottleError(&types.ProvisionedThroughputExceededException{}))
	require.False(t, isDynamoDBThrottleError(errors.New("access denied")))
}

func TestBatchWriteBackoffDuration_IncreasesAndCaps(t *testing.T) {
	first := batchWriteBackoffDuration(1)
	second := batchWriteBackoffDuration(2)
	large := batchWriteBackoffDuration(20)

	require.GreaterOrEqual(t, first, batchWriteBackoffMin)
	require.GreaterOrEqual(t, second, first)
	require.LessOrEqual(t, large, batchWriteBackoffMax+(batchWriteBackoffMax/4))
}

func TestSleepWithContext_RespectsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sleepWithContext(ctx, time.Second)
	require.ErrorIs(t, err, context.Canceled)
}
