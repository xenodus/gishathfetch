package ckprices

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"mtg-price-checker-sg/gateway/cardkingdom"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	batchGetLimit            = 100
	batchWriteLimit          = 25
	priceChangeIndexName     = "priceChangePercent-index"
	priceChangeUsdIndexName  = "priceChangeUsd-index"
	priceChangeIndexPKValue  = "CURRENT"
	syncMetadataKey          = "__sync__"
	syncMetadataLabel        = "CK price sync metadata"
)

type dynamoRecord struct {
	NameKey            string   `dynamodbav:"nameKey"`
	CardName           string   `dynamodbav:"cardName"`
	Edition            string   `dynamodbav:"edition"`
	PriceUsd           float64  `dynamodbav:"priceUsd"`
	PreviousPriceUsd   *float64 `dynamodbav:"previousPriceUsd,omitempty"`
	PriceChangePercent *int     `dynamodbav:"priceChangePercent,omitempty"`
	PriceChangeUsd     *float64 `dynamodbav:"priceChangeUsd,omitempty"`
	PriceChangeIndexPK *string `dynamodbav:"priceChangeIndexPK,omitempty"`
	URL                string  `dynamodbav:"url"`
	IsFoil             bool    `dynamodbav:"isFoil"`
	UpdatedAt          string  `dynamodbav:"updatedAt"`
	SyncedAt           string  `dynamodbav:"syncedAt"`
}

type syncMetadataRecord struct {
	NameKey      string `dynamodbav:"nameKey"`
	CardName     string `dynamodbav:"cardName"`
	SyncedAt     string `dynamodbav:"syncedAt"`
	ListingCount int    `dynamodbav:"listingCount"`
}

type DynamoDBStore struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBStore(ctx context.Context) (*DynamoDBStore, error) {
	tableName := strings.TrimSpace(os.Getenv(config.CKDynamoDBTableEnv))
	if tableName == "" {
		return nil, fmt.Errorf("ckprices: %s is not set", config.CKDynamoDBTableEnv)
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(config.AWSRegion))
	if err != nil {
		return nil, err
	}

	return &DynamoDBStore{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}, nil
}

func (s *DynamoDBStore) GetByNameKey(ctx context.Context, nameKey string) (*cardkingdom.Listing, error) {
	output, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"nameKey": &types.AttributeValueMemberS{Value: nameKey},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(output.Item) == 0 {
		return nil, nil
	}

	var record dynamoRecord
	if err := attributevalue.UnmarshalMap(output.Item, &record); err != nil {
		return nil, err
	}

	listing, ok := listingFromRecord(record)
	if !ok {
		return nil, nil
	}
	return &listing, nil
}

func (s *DynamoDBStore) GetPriceChangesByPercent(ctx context.Context, ascending bool, limit int) ([]PriceChangeListing, error) {
	if limit <= 0 {
		limit = PriceChangeRankingLimit
	}

	listings, err := s.queryPriceChangesByPercent(ctx, ascending, limit)
	if err == nil {
		return listings, nil
	}
	if !isMissingPriceChangeIndex(err) {
		return nil, err
	}

	scanned, scanErr := s.scanPriceChangeListings(ctx)
	if scanErr != nil {
		return nil, scanErr
	}
	return priceChangesByPercentFromListings(scanned, ascending, limit), nil
}

func (s *DynamoDBStore) GetPriceChangesByUsd(ctx context.Context, ascending bool, limit int) ([]PriceChangeListing, error) {
	if limit <= 0 {
		limit = PriceChangeRankingLimit
	}

	listings, err := s.queryPriceChangesByUsd(ctx, ascending, limit)
	if err == nil {
		return listings, nil
	}
	if !isMissingPriceChangeIndex(err) {
		return nil, err
	}

	scanned, scanErr := s.scanPriceChangeListingsByUsd(ctx)
	if scanErr != nil {
		return nil, scanErr
	}
	return priceChangesByUsdFromListings(scanned, ascending, limit), nil
}

func (s *DynamoDBStore) GetTopBottomPriceChanges(ctx context.Context) (*TopBottomPriceChanges, error) {
	top, err := s.GetPriceChangesByUsd(ctx, false, PriceChangeRankingLimit)
	if err != nil {
		return nil, err
	}
	bottom, err := s.GetPriceChangesByUsd(ctx, true, PriceChangeRankingLimit)
	if err != nil {
		return nil, err
	}
	// Top must only contain price increases and Bottom only price drops. When
	// fewer than PriceChangeRankingLimit listings moved in a direction, the raw
	// rankings would otherwise spill over into the opposite sign.
	return &TopBottomPriceChanges{
		Top:    filterPriceChangesByUsdSign(top, true),
		Bottom: filterPriceChangesByUsdSign(bottom, false),
	}, nil
}

func (s *DynamoDBStore) queryPriceChangesByPercent(ctx context.Context, ascending bool, limit int) ([]PriceChangeListing, error) {
	output, err := s.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		IndexName:              aws.String(priceChangeIndexName),
		KeyConditionExpression: aws.String("priceChangeIndexPK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: priceChangeIndexPKValue},
		},
		ScanIndexForward: aws.Bool(ascending),
		Limit:            aws.Int32(int32(limit)),
	})
	if err != nil {
		return nil, err
	}

	return priceChangeListingsFromItems(output.Items)
}

func (s *DynamoDBStore) queryPriceChangesByUsd(ctx context.Context, ascending bool, limit int) ([]PriceChangeListing, error) {
	output, err := s.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		IndexName:              aws.String(priceChangeUsdIndexName),
		KeyConditionExpression: aws.String("priceChangeIndexPK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: priceChangeIndexPKValue},
		},
		ScanIndexForward: aws.Bool(ascending),
		Limit:            aws.Int32(int32(limit)),
	})
	if err != nil {
		return nil, err
	}

	return priceChangeListingsFromItemsByUsd(output.Items)
}

func (s *DynamoDBStore) scanPriceChangeListings(ctx context.Context) ([]PriceChangeListing, error) {
	return s.scanPriceChangeListingsWithFilter(ctx, priceChangeListingFromRecord)
}

func (s *DynamoDBStore) scanPriceChangeListingsByUsd(ctx context.Context) ([]PriceChangeListing, error) {
	return s.scanPriceChangeListingsWithFilter(ctx, priceChangeListingFromRecordByUsd)
}

func (s *DynamoDBStore) scanPriceChangeListingsWithFilter(
	ctx context.Context,
	fromRecord func(dynamoRecord) (PriceChangeListing, bool),
) ([]PriceChangeListing, error) {
	listings := make([]PriceChangeListing, 0)
	var exclusiveStartKey map[string]types.AttributeValue

	for {
		output, err := s.client.Scan(ctx, &dynamodb.ScanInput{
			TableName:         aws.String(s.tableName),
			ExclusiveStartKey: exclusiveStartKey,
		})
		if err != nil {
			return nil, err
		}

		batch, err := priceChangeListingsFromItemsWithFilter(output.Items, fromRecord)
		if err != nil {
			return nil, err
		}
		listings = append(listings, batch...)

		if len(output.LastEvaluatedKey) == 0 {
			break
		}
		exclusiveStartKey = output.LastEvaluatedKey
	}

	return listings, nil
}

func (s *DynamoDBStore) PutAll(ctx context.Context, listings map[string]cardkingdom.Listing) (string, error) {
	nameKeys := make([]string, 0, len(listings))
	for nameKey := range listings {
		nameKeys = append(nameKeys, nameKey)
	}

	existing, err := s.batchGetExisting(ctx, nameKeys)
	if err != nil {
		return "", err
	}
	listings = listingsWithPriceChange(existing, listings)

	syncedAt := time.Now().UTC().Format(time.RFC3339)
	writeRequests := make([]types.WriteRequest, 0, len(listings)+1)
	for nameKey, listing := range listings {
		record := dynamoRecordFromListing(nameKey, listing, syncedAt)
		item, err := attributevalue.MarshalMap(record)
		if err != nil {
			return "", err
		}
		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{Item: item},
		})
	}

	metadataItem, err := attributevalue.MarshalMap(syncMetadataRecord{
		NameKey:      syncMetadataKey,
		CardName:     syncMetadataLabel,
		SyncedAt:     syncedAt,
		ListingCount: len(listings),
	})
	if err != nil {
		return "", err
	}
	writeRequests = append(writeRequests, types.WriteRequest{
		PutRequest: &types.PutRequest{Item: metadataItem},
	})

	for start := 0; start < len(writeRequests); start += batchWriteLimit {
		end := min(start+batchWriteLimit, len(writeRequests))
		batch := writeRequests[start:end]
		if err := s.writeBatch(ctx, batch); err != nil {
			return "", err
		}
	}

	return syncedAt, nil
}

func dynamoRecordFromListing(nameKey string, listing cardkingdom.Listing, syncedAt string) dynamoRecord {
	record := dynamoRecord{
		NameKey:            nameKey,
		CardName:           listing.CardName,
		Edition:            listing.Edition,
		PriceUsd:           listing.PriceUsd,
		PreviousPriceUsd:   listing.PreviousPriceUsd,
		PriceChangePercent: listing.PriceChangePercent,
		PriceChangeUsd:     listing.PriceChangeUsd,
		URL:                listing.URL,
		IsFoil:             listing.IsFoil,
		UpdatedAt:          listing.UpdatedAt,
		SyncedAt:           syncedAt,
	}
	if listing.PriceChangeUsd != nil || listing.PriceChangePercent != nil {
		indexPK := priceChangeIndexPKValue
		record.PriceChangeIndexPK = &indexPK
	}
	return record
}

func listingFromRecord(record dynamoRecord) (cardkingdom.Listing, bool) {
	if record.NameKey == syncMetadataKey {
		return cardkingdom.Listing{}, false
	}
	return cardkingdom.Listing{
		CardName:           record.CardName,
		Edition:            record.Edition,
		PriceUsd:           record.PriceUsd,
		PreviousPriceUsd:   record.PreviousPriceUsd,
		PriceChangePercent: record.PriceChangePercent,
		PriceChangeUsd:     record.PriceChangeUsd,
		URL:                record.URL,
		IsFoil:             record.IsFoil,
		UpdatedAt:          record.UpdatedAt,
		SyncedAt:           record.SyncedAt,
	}, true
}

func priceChangeListingFromRecord(record dynamoRecord) (PriceChangeListing, bool) {
	if record.NameKey == syncMetadataKey || record.PriceChangePercent == nil {
		return PriceChangeListing{}, false
	}
	listing, ok := listingFromRecord(record)
	if !ok {
		return PriceChangeListing{}, false
	}
	return PriceChangeListing{
		NameKey: record.NameKey,
		Listing: listing,
	}, true
}

func priceChangeListingFromRecordByUsd(record dynamoRecord) (PriceChangeListing, bool) {
	if record.NameKey == syncMetadataKey || record.PriceChangeUsd == nil {
		return PriceChangeListing{}, false
	}
	listing, ok := listingFromRecord(record)
	if !ok {
		return PriceChangeListing{}, false
	}
	return PriceChangeListing{
		NameKey: record.NameKey,
		Listing: listing,
	}, true
}

func priceChangeListingsFromItems(items []map[string]types.AttributeValue) ([]PriceChangeListing, error) {
	return priceChangeListingsFromItemsWithFilter(items, priceChangeListingFromRecord)
}

func priceChangeListingsFromItemsByUsd(items []map[string]types.AttributeValue) ([]PriceChangeListing, error) {
	return priceChangeListingsFromItemsWithFilter(items, priceChangeListingFromRecordByUsd)
}

func priceChangeListingsFromItemsWithFilter(
	items []map[string]types.AttributeValue,
	fromRecord func(dynamoRecord) (PriceChangeListing, bool),
) ([]PriceChangeListing, error) {
	listings := make([]PriceChangeListing, 0, len(items))
	for _, item := range items {
		var record dynamoRecord
		if err := attributevalue.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if listing, ok := fromRecord(record); ok {
			listings = append(listings, listing)
		}
	}
	return listings, nil
}

func isMissingPriceChangeIndex(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "index") &&
		(strings.Contains(message, "not found") ||
			strings.Contains(message, "does not have the specified index") ||
			strings.Contains(message, "validationexception"))
}

func (s *DynamoDBStore) batchGetExisting(ctx context.Context, nameKeys []string) (map[string]dynamoRecord, error) {
	existing := make(map[string]dynamoRecord, len(nameKeys))
	if len(nameKeys) == 0 {
		return existing, nil
	}

	for start := 0; start < len(nameKeys); start += batchGetLimit {
		end := min(start+batchGetLimit, len(nameKeys))
		batch := nameKeys[start:end]

		keys := make([]map[string]types.AttributeValue, len(batch))
		for i, nameKey := range batch {
			keys[i] = map[string]types.AttributeValue{
				"nameKey": &types.AttributeValueMemberS{Value: nameKey},
			}
		}

		pending := map[string]types.KeysAndAttributes{
			s.tableName: {Keys: keys},
		}
		for len(pending) > 0 {
			output, err := s.client.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
				RequestItems: pending,
			})
			if err != nil {
				return nil, err
			}

			for _, item := range output.Responses[s.tableName] {
				var record dynamoRecord
				if err := attributevalue.UnmarshalMap(item, &record); err != nil {
					return nil, err
				}
				if record.NameKey == syncMetadataKey {
					continue
				}
				existing[record.NameKey] = record
			}

			if len(output.UnprocessedKeys) == 0 {
				break
			}
			pending = output.UnprocessedKeys
		}
	}

	return existing, nil
}

func (s *DynamoDBStore) writeBatch(ctx context.Context, batch []types.WriteRequest) error {
	pending := map[string][]types.WriteRequest{
		s.tableName: batch,
	}

	for len(pending) > 0 {
		output, err := s.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: pending,
		})
		if err != nil {
			return err
		}
		if len(output.UnprocessedItems) == 0 {
			return nil
		}
		pending = output.UnprocessedItems
	}

	return nil
}
