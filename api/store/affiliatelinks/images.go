package affiliatelinks

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const maxImageBytes = 2 * 1024 * 1024

var allowedImageContentTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

type ImageUploader struct {
	client objectWriter
	bucket string
	prefix string
}

type objectWriter interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

var loadAWSConfig = func(ctx context.Context) (aws.Config, error) {
	return awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(config.AWSRegion))
}

func NewImageUploader(ctx context.Context) (*ImageUploader, error) {
	bucket := strings.TrimSpace(os.Getenv(config.AffiliateImagesS3BucketEnv))
	if bucket == "" {
		bucket = config.AffiliateImagesS3DefaultBucket
	}

	prefix := strings.Trim(strings.TrimSpace(os.Getenv(config.AffiliateImagesS3PrefixEnv)), "/")
	if prefix == "" {
		prefix = config.AffiliateImagesS3DefaultPrefix
	}

	cfg, err := loadAWSConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &ImageUploader{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
		prefix: prefix,
	}, nil
}

func (u *ImageUploader) Upload(ctx context.Context, data []byte, contentType string) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("affiliatelinks: image data is required")
	}
	if len(data) > maxImageBytes {
		return "", fmt.Errorf("affiliatelinks: image exceeds %d byte limit", maxImageBytes)
	}

	ext, ok := allowedImageContentTypes[strings.ToLower(strings.TrimSpace(contentType))]
	if !ok {
		return "", fmt.Errorf("affiliatelinks: unsupported image content type %q", contentType)
	}

	key := path.Join(u.prefix, newImageObjectID()+ext)
	_, err := u.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("affiliatelinks: upload image: %w", err)
	}

	return fmt.Sprintf("https://%s/%s", u.bucket, key), nil
}
