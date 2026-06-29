package affiliatelinks

import (
	"context"
	"fmt"
	"os"
	"strings"

	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBStore struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBStore(ctx context.Context) (*DynamoDBStore, error) {
	tableName := strings.TrimSpace(os.Getenv(config.AffiliateLinksDynamoDBTableEnv))
	if tableName == "" {
		return nil, fmt.Errorf("affiliatelinks: %s is not set", config.AffiliateLinksDynamoDBTableEnv)
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

func (s *DynamoDBStore) ListAll(ctx context.Context) ([]Link, error) {
	var links []Link
	var lastEvaluatedKey map[string]types.AttributeValue

	for {
		output, err := s.client.Scan(ctx, &dynamodb.ScanInput{
			TableName:         aws.String(s.tableName),
			ExclusiveStartKey: lastEvaluatedKey,
		})
		if err != nil {
			return nil, err
		}

		var page []Link
		if err := attributevalue.UnmarshalListOfMaps(output.Items, &page); err != nil {
			return nil, err
		}
		links = append(links, page...)

		if len(output.LastEvaluatedKey) == 0 {
			break
		}
		lastEvaluatedKey = output.LastEvaluatedKey
	}

	return links, nil
}

func (s *DynamoDBStore) GetByID(ctx context.Context, id string) (*Link, error) {
	output, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(output.Item) == 0 {
		return nil, nil
	}

	var link Link
	if err := attributevalue.UnmarshalMap(output.Item, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

func (s *DynamoDBStore) Put(ctx context.Context, link Link) error {
	item, err := attributevalue.MarshalMap(link)
	if err != nil {
		return err
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})
	return err
}

func (s *DynamoDBStore) Delete(ctx context.Context, id string) error {
	_, err := s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	return err
}
