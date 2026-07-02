package adminlogin

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

const (
	rateKeyPrefix = "rate#"
	logKeyPrefix  = "log#"
)

type rateRecord struct {
	PK          string `dynamodbav:"pk"`
	FailCount   int    `dynamodbav:"failCount"`
	WindowStart string `dynamodbav:"windowStart"`
	LockedUntil string `dynamodbav:"lockedUntil,omitempty"`
	TTL         int64  `dynamodbav:"ttl"`
}

type attemptRecord struct {
	PK        string `dynamodbav:"pk"`
	IP        string `dynamodbav:"ip"`
	Username  string `dynamodbav:"username"`
	Success   bool   `dynamodbav:"success"`
	Blocked   bool   `dynamodbav:"blocked"`
	UserAgent string `dynamodbav:"userAgent,omitempty"`
	CreatedAt string `dynamodbav:"createdAt"`
	TTL       int64  `dynamodbav:"ttl"`
}

type DynamoDBStore struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBStore(ctx context.Context) (*DynamoDBStore, error) {
	tableName := strings.TrimSpace(os.Getenv(config.AdminLoginDynamoDBTableEnv))
	if tableName == "" {
		return nil, fmt.Errorf("adminlogin: %s is not set", config.AdminLoginDynamoDBTableEnv)
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

func (s *DynamoDBStore) CheckLockout(ctx context.Context, ip, username string, now time.Time) (Lockout, error) {
	ipLockout, err := s.checkKeyLockout(ctx, rateKey("ip", ip), now)
	if err != nil {
		return Lockout{}, err
	}
	if ipLockout.Locked {
		return ipLockout, nil
	}

	return s.checkKeyLockout(ctx, rateKey("user", username), now)
}

func (s *DynamoDBStore) RecordAttempt(ctx context.Context, attempt Attempt, retention time.Duration) error {
	if attempt.ID == "" {
		attempt.ID = uuid.NewString()
	}
	if attempt.CreatedAt.IsZero() {
		attempt.CreatedAt = time.Now().UTC()
	}

	record := attemptRecord{
		PK:        logKeyPrefix + attempt.ID,
		IP:        attempt.IP,
		Username:  attempt.Username,
		Success:   attempt.Success,
		Blocked:   attempt.Blocked,
		UserAgent: attempt.UserAgent,
		CreatedAt: attempt.CreatedAt.UTC().Format(time.RFC3339),
		TTL:       attempt.CreatedAt.UTC().Add(retention).Unix(),
	}

	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return err
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})
	return err
}

func (s *DynamoDBStore) RecordFailure(
	ctx context.Context,
	ip, username string,
	now time.Time,
	limits RateLimits,
) error {
	if err := s.incrementFailure(ctx, rateKey("ip", ip), now, limits.MaxFailuresPerIP, limits.IPWindow, limits.IPLockout); err != nil {
		return err
	}
	return s.incrementFailure(ctx, rateKey("user", username), now, limits.MaxFailuresPerUser, limits.UserWindow, limits.UserLockout)
}

func (s *DynamoDBStore) ClearFailures(ctx context.Context, ip, username string) error {
	if err := s.deleteRateRecord(ctx, rateKey("ip", ip)); err != nil {
		return err
	}
	return s.deleteRateRecord(ctx, rateKey("user", username))
}

func (s *DynamoDBStore) checkKeyLockout(ctx context.Context, pk string, now time.Time) (Lockout, error) {
	record, err := s.getRateRecord(ctx, pk)
	if err != nil {
		return Lockout{}, err
	}
	if record == nil {
		return Lockout{}, nil
	}

	lockedUntil, err := time.Parse(time.RFC3339, record.LockedUntil)
	if err != nil || !lockedUntil.After(now) {
		return Lockout{}, nil
	}

	return Lockout{
		Locked:     true,
		RetryAfter: lockedUntil.Sub(now),
	}, nil
}

func (s *DynamoDBStore) incrementFailure(
	ctx context.Context,
	pk string,
	now time.Time,
	maxFailures int,
	window time.Duration,
	lockout time.Duration,
) error {
	record, err := s.getRateRecord(ctx, pk)
	if err != nil {
		return err
	}

	current := rateRecord{
		PK:          pk,
		FailCount:   0,
		WindowStart: now.UTC().Format(time.RFC3339),
	}
	if record != nil {
		current = *record
	}

	windowStart, err := time.Parse(time.RFC3339, current.WindowStart)
	if err != nil || now.Sub(windowStart) > window {
		current.FailCount = 0
		current.WindowStart = now.UTC().Format(time.RFC3339)
		current.LockedUntil = ""
	}

	current.FailCount++
	if current.FailCount >= maxFailures {
		lockedUntil := now.Add(lockout)
		current.LockedUntil = lockedUntil.UTC().Format(time.RFC3339)
		current.TTL = lockedUntil.Add(time.Hour).Unix()
	} else {
		current.TTL = now.Add(window).Add(time.Hour).Unix()
	}

	item, err := attributevalue.MarshalMap(current)
	if err != nil {
		return err
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})
	return err
}

func (s *DynamoDBStore) getRateRecord(ctx context.Context, pk string) (*rateRecord, error) {
	output, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: pk},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(output.Item) == 0 {
		return nil, nil
	}

	var record rateRecord
	if err := attributevalue.UnmarshalMap(output.Item, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *DynamoDBStore) deleteRateRecord(ctx context.Context, pk string) error {
	_, err := s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: pk},
		},
	})
	return err
}

func rateKey(kind, value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		normalized = "unknown"
	}
	return rateKeyPrefix + kind + "#" + normalized
}
