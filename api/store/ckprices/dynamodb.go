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
	batchWriteLimit   = 25
	syncMetadataKey   = "__sync__"
	syncMetadataLabel = "CK price sync metadata"
)

type dynamoRecord struct {
	NameKey   string  `dynamodbav:"nameKey"`
	CardName  string  `dynamodbav:"cardName"`
	Edition   string  `dynamodbav:"edition"`
	PriceUsd  float64 `dynamodbav:"priceUsd"`
	URL       string  `dynamodbav:"url"`
	Quantity  int     `dynamodbav:"quantity"`
	IsFoil    bool    `dynamodbav:"isFoil"`
	UpdatedAt string  `dynamodbav:"updatedAt"`
	SyncedAt  string  `dynamodbav:"syncedAt"`
}

type syncMetadataRecord struct {
	NameKey       string `dynamodbav:"nameKey"`
	CardName      string `dynamodbav:"cardName"`
	SyncedAt      string `dynamodbav:"syncedAt"`
	ListingCount  int    `dynamodbav:"listingCount"`
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

	return &cardkingdom.Listing{
		CardName:  record.CardName,
		Edition:   record.Edition,
		PriceUsd:  record.PriceUsd,
		URL:       record.URL,
		Quantity:  record.Quantity,
		IsFoil:    record.IsFoil,
		UpdatedAt: record.UpdatedAt,
		SyncedAt:  record.SyncedAt,
	}, nil
}

func (s *DynamoDBStore) PutAll(ctx context.Context, listings map[string]cardkingdom.Listing) (string, error) {
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
		end := start + batchWriteLimit
		if end > len(writeRequests) {
			end = len(writeRequests)
		}
		batch := writeRequests[start:end]
		if err := s.writeBatch(ctx, batch); err != nil {
			return "", err
		}
	}

	return syncedAt, nil
}

func dynamoRecordFromListing(nameKey string, listing cardkingdom.Listing, syncedAt string) dynamoRecord {
	return dynamoRecord{
		NameKey:   nameKey,
		CardName:  listing.CardName,
		Edition:   listing.Edition,
		PriceUsd:  listing.PriceUsd,
		URL:       listing.URL,
		Quantity:  listing.Quantity,
		IsFoil:    listing.IsFoil,
		UpdatedAt: listing.UpdatedAt,
		SyncedAt:  syncedAt,
	}
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
