package analyticsreport

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"mtg-price-checker-sg/controller/analyticskeywords"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type mockS3Client struct {
	objects map[string][]byte
}

func (m *mockS3Client) PutObject(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if m.objects == nil {
		m.objects = make(map[string][]byte)
	}

	body, err := io.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}
	m.objects[aws.ToString(input.Key)] = body
	return &s3.PutObjectOutput{}, nil
}

func TestS3Writer_WriteUploadsLatestAndDatedObjects(t *testing.T) {
	mockClient := &mockS3Client{}
	writer := &S3Writer{
		client: mockClient,
		bucket: "analytics-bucket",
		prefix: "analytics/top-search-keywords",
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

	if len(mockClient.objects) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(mockClient.objects))
	}

	latestKey := "analytics/top-search-keywords/latest.json"
	datedKey := "analytics/top-search-keywords/2026-06-28.json"
	if _, ok := mockClient.objects[latestKey]; !ok {
		t.Fatalf("missing latest object")
	}
	if _, ok := mockClient.objects[datedKey]; !ok {
		t.Fatalf("missing dated object")
	}

	var decoded analyticskeywords.Report
	if err := json.Unmarshal(mockClient.objects[latestKey], &decoded); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if decoded.PropertyID != "123456789" {
		t.Fatalf("unexpected decoded report: %+v", decoded)
	}
}

func TestParseGeneratedAtDate(t *testing.T) {
	got, err := ParseGeneratedAtDate("2026-06-28T01:02:03Z")
	if err != nil {
		t.Fatalf("parse generated at: %v", err)
	}
	if got != "2026-06-28" {
		t.Fatalf("expected 2026-06-28, got %s", got)
	}
}
