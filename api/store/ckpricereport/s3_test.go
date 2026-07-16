package ckpricereport

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"mtg-price-checker-sg/gateway/cardkingdom"
	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/store/ckprices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type mockS3Object struct {
	body         []byte
	contentType  string
	cacheControl string
}

type mockS3Client struct {
	objects map[string]mockS3Object
}

func (m *mockS3Client) GetObject(_ context.Context, input *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	key := aws.ToString(input.Key)
	object, ok := m.objects[key]
	if !ok {
		return nil, &types.NoSuchKey{}
	}
	return &s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewReader(object.body)),
	}, nil
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

func TestNewS3WriterUsesFrontendExportPath(t *testing.T) {
	writer, err := NewS3Writer(context.Background())
	if err != nil {
		t.Fatalf("new s3 writer: %v", err)
	}
	if writer.bucket != config.CKPriceChangesS3Bucket {
		t.Fatalf("expected bucket %q, got %q", config.CKPriceChangesS3Bucket, writer.bucket)
	}
	if writer.prefix != config.CKPriceChangesS3KeyPrefix {
		t.Fatalf("expected prefix %q, got %q", config.CKPriceChangesS3KeyPrefix, writer.prefix)
	}
}

func TestS3Writer_WriteUploadsLatestObject(t *testing.T) {
	mockClient := &mockS3Client{}
	writer := &S3Writer{
		client: mockClient,
		bucket: config.CKPriceChangesS3Bucket,
		prefix: config.CKPriceChangesS3KeyPrefix,
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

	latestKey := config.CKPriceChangesS3KeyPrefix + "/latest.json"
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

func TestS3Writer_ReadLatestReturnsStoredReport(t *testing.T) {
	increase := 1.0
	report := &Report{
		GeneratedAt:  "2026-07-16T17:35:00Z",
		RankingLimit: 20,
		Top: []ckprices.PriceChangeListing{{
			NameKey: "bolt",
			Listing: cardkingdom.Listing{PriceChangeUsd: &increase},
		}},
	}
	payload, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}

	mockClient := &mockS3Client{
		objects: map[string]mockS3Object{
			config.CKPriceChangesS3KeyPrefix + "/latest.json": {body: payload},
		},
	}
	writer := &S3Writer{
		client: mockClient,
		bucket: config.CKPriceChangesS3Bucket,
		prefix: config.CKPriceChangesS3KeyPrefix,
	}

	got, err := writer.ReadLatest(context.Background())
	if err != nil {
		t.Fatalf("read latest: %v", err)
	}
	if got.GeneratedAt != report.GeneratedAt {
		t.Fatalf("generatedAt = %q, want %q", got.GeneratedAt, report.GeneratedAt)
	}
	if len(got.Top) != 1 || got.Top[0].NameKey != "bolt" {
		t.Fatalf("unexpected top rankings: %+v", got.Top)
	}
}

func TestS3Writer_ReadLatestMissingObject(t *testing.T) {
	writer := &S3Writer{
		client: &mockS3Client{},
		bucket: config.CKPriceChangesS3Bucket,
		prefix: config.CKPriceChangesS3KeyPrefix,
	}

	_, err := writer.ReadLatest(context.Background())
	if err != ErrLatestReportNotFound {
		t.Fatalf("err = %v, want %v", err, ErrLatestReportNotFound)
	}
}
