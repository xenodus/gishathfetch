package analyticsreport

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"mtg-price-checker-sg/controller/analyticskeywords"
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
	t.Setenv(config.AnalyticsS3BucketEnv, "")
	t.Setenv(config.AnalyticsS3KeyPrefixEnv, "")

	writer, err := NewS3Writer(context.Background())
	if err != nil {
		t.Fatalf("new s3 writer: %v", err)
	}
	if writer.bucket != config.AnalyticsS3DefaultBucket {
		t.Fatalf("expected bucket %q, got %q", config.AnalyticsS3DefaultBucket, writer.bucket)
	}
	if writer.prefix != config.AnalyticsS3DefaultKeyPrefix {
		t.Fatalf("expected prefix %q, got %q", config.AnalyticsS3DefaultKeyPrefix, writer.prefix)
	}
}

func TestS3Writer_WriteUploadsLatestObject(t *testing.T) {
	mockClient := &mockS3Client{}
	writer := &S3Writer{
		client: mockClient,
		bucket: config.AnalyticsS3DefaultBucket,
		prefix: config.AnalyticsS3DefaultKeyPrefix,
	}

	report := &analyticskeywords.Report{
		GeneratedAt: "2026-06-28T01:02:03Z",
		PropertyID:  "123456789",
		EventName:   "search",
		Periods: map[string]analyticskeywords.PeriodReport{
			"last7Days": {
				Keywords: []analyticskeywords.KeywordCount{{Term: "Opt", Count: 3}},
			},
		},
	}

	if err := writer.Write(context.Background(), report); err != nil {
		t.Fatalf("write report: %v", err)
	}

	if len(mockClient.objects) != 1 {
		t.Fatalf("expected 1 object, got %d", len(mockClient.objects))
	}

	latestKey := config.AnalyticsS3DefaultKeyPrefix + "/latest.json"
	object, ok := mockClient.objects[latestKey]
	if !ok {
		t.Fatalf("missing latest object")
	}
	if object.contentType != "application/json" {
		t.Fatalf("expected application/json content type, got %q", object.contentType)
	}
	if object.cacheControl != config.AnalyticsLatestJSONCacheControl {
		t.Fatalf("expected cache control %q, got %q", config.AnalyticsLatestJSONCacheControl, object.cacheControl)
	}

	var decoded analyticskeywords.Report
	if err := json.Unmarshal(object.body, &decoded); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if decoded.PropertyID != "123456789" {
		t.Fatalf("unexpected decoded report: %+v", decoded)
	}
}
