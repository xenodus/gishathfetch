package ckpricereport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// Writer persists CK price change reports to S3.
type Writer interface {
	Write(ctx context.Context, report *Report) error
}

type objectClient interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// ErrLatestReportNotFound is returned when latest.json has not been written yet.
var ErrLatestReportNotFound = errors.New("ckpricereport: latest report not found")

// S3Writer uploads JSON reports to a configured bucket and key prefix.
type S3Writer struct {
	client objectClient
	bucket string
	prefix string
}

var loadAWSConfig = func(ctx context.Context) (aws.Config, error) {
	return awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(config.AWSRegion))
}

// NewS3Writer builds an S3 writer for the frontend CK price change export path.
func NewS3Writer(ctx context.Context) (*S3Writer, error) {
	cfg, err := loadAWSConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &S3Writer{
		client: s3.NewFromConfig(cfg),
		bucket: config.CKPriceChangesS3Bucket,
		prefix: config.CKPriceChangesS3KeyPrefix,
	}, nil
}

func (w *S3Writer) Write(ctx context.Context, report *Report) error {
	if report == nil {
		return fmt.Errorf("ckpricereport: report is required")
	}

	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	key := w.objectKey("latest.json")
	_, err = w.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(w.bucket),
		Key:          aws.String(key),
		Body:         bytes.NewReader(payload),
		ContentType:  aws.String("application/json"),
		CacheControl: aws.String(config.CKPriceChangesLatestJSONCacheControl),
	})
	if err != nil {
		return fmt.Errorf("ckpricereport: put s3://%s/%s: %w", w.bucket, key, err)
	}

	return nil
}

// ReadLatest loads the current CK price change export from S3.
func (w *S3Writer) ReadLatest(ctx context.Context) (*Report, error) {
	key := w.objectKey("latest.json")
	output, err := w.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(w.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isS3NotFound(err) {
			return nil, ErrLatestReportNotFound
		}
		return nil, fmt.Errorf("ckpricereport: get s3://%s/%s: %w", w.bucket, key, err)
	}
	defer output.Body.Close()

	body, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("ckpricereport: read s3://%s/%s: %w", w.bucket, key, err)
	}

	var report Report
	if err := json.Unmarshal(body, &report); err != nil {
		return nil, fmt.Errorf("ckpricereport: decode s3://%s/%s: %w", w.bucket, key, err)
	}
	return &report, nil
}

func isS3NotFound(err error) bool {
	var noSuchKey *types.NoSuchKey
	if errors.As(err, &noSuchKey) {
		return true
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NoSuchKey" {
		return true
	}
	return false
}

func (w *S3Writer) objectKey(name string) string {
	return w.prefix + "/" + name
}
