package analyticsreport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"mtg-price-checker-sg/controller/analyticskeywords"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Writer persists analytics keyword reports to S3.
type Writer interface {
	Write(ctx context.Context, report *analyticskeywords.Report) error
}

type objectWriter interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// S3Writer uploads JSON reports to a configured bucket and key prefix.
type S3Writer struct {
	client objectWriter
	bucket string
	prefix string
}

var loadAWSConfig = func(ctx context.Context) (aws.Config, error) {
	return awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(config.AWSRegion))
}

// NewS3Writer builds an S3 writer from environment configuration.
func NewS3Writer(ctx context.Context) (*S3Writer, error) {
	bucket := strings.TrimSpace(os.Getenv(config.AnalyticsS3BucketEnv))
	if bucket == "" {
		return nil, fmt.Errorf("analyticsreport: %s is not set", config.AnalyticsS3BucketEnv)
	}

	prefix := strings.Trim(strings.TrimSpace(os.Getenv(config.AnalyticsS3KeyPrefixEnv)), "/")
	if prefix == "" {
		prefix = "analytics/top-search-keywords"
	}

	cfg, err := loadAWSConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &S3Writer{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
		prefix: prefix,
	}, nil
}

func (w *S3Writer) Write(ctx context.Context, report *analyticskeywords.Report) error {
	if report == nil {
		return fmt.Errorf("analyticsreport: report is required")
	}

	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	key := w.objectKey("latest.json")
	_, err = w.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(w.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(payload),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("analyticsreport: put s3://%s/%s: %w", w.bucket, key, err)
	}

	return nil
}

func (w *S3Writer) objectKey(name string) string {
	return w.prefix + "/" + name
}
