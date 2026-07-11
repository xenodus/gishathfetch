package ckpricereport

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type mockS3Object struct {
	body         []byte
	contentType  string
	cacheControl string
}

type mockS3Client struct {
	objects map[string]mockS3Object
}

func (m *mockS3Client) PutObject(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if m.objects == nil {
		m.objects = make(map[string]mockS3Object)
	}

	body, err := io.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}
	m.objects[aws.ToString(input.Key)] = mockS3Object{
		body:         body,
		contentType:  aws.ToString(input.ContentType),
		cacheControl: aws.ToString(input.CacheControl),
	}
	return &s3.PutObjectOutput{}, nil
}

func TestNewS3WriterDefaultsToFrontendBucket(t *testing.T) {
	t.Setenv(config.CKPriceChangesS3BucketEnv, "")
	t.Setenv(config.CKPriceChangesS3KeyPrefixEnv, "")

	writer, err := NewS3Writer(context.Background())
	if err != nil {
		t.Fatalf("new s3 writer: %v", err)
	}
	if writer.bucket != config.CKPriceChangesS3DefaultBucket {
		t.Fatalf("expected bucket %q, got %q", config.CKPriceChangesS3DefaultBucket, writer.bucket)
	}
	if writer.prefix != config.CKPriceChangesS3DefaultKeyPrefix {
		t.Fatalf("expected prefix %q, got %q", config.CKPriceChangesS3DefaultKeyPrefix, writer.prefix)
	}
}

func TestS3Writer_WriteUploadsLatestObject(t *testing.T) {
	mockClient := &mockS3Client{}
	writer := &S3Writer{
		client: mockClient,
		bucket: config.CKPriceChangesS3DefaultBucket,
		prefix: config.CKPriceChangesS3DefaultKeyPrefix,
	}

	report := &Report{
		GeneratedAt:  "2026-07-11T12:00:00Z",
		SyncedAt:     "2026-07-11T00:00:00Z",
		RankingLimit: 20,
		Top:          nil,
		Bottom:       nil,
	}

	if err := writer.Write(context.Background(), report); err != nil {
		t.Fatalf("write report: %v", err)
	}

	if len(mockClient.objects) != 1 {
		t.Fatalf("expected 1 object, got %d", len(mockClient.objects))
	}

	latestKey := config.CKPriceChangesS3DefaultKeyPrefix + "/latest.json"
	object, ok := mockClient.objects[latestKey]
	if !ok {
		t.Fatalf("missing latest object")
	}
	if object.contentType != "application/json" {
		t.Fatalf("expected application/json content type, got %q", object.contentType)
	}
	if object.cacheControl != config.CKPriceChangesLatestJSONCacheControl {
		t.Fatalf("expected cache control %q, got %q", config.CKPriceChangesLatestJSONCacheControl, object.cacheControl)
	}

	var decoded Report
	if err := json.Unmarshal(object.body, &decoded); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if decoded.GeneratedAt != report.GeneratedAt {
		t.Fatalf("unexpected decoded report: %+v", decoded)
	}
}
