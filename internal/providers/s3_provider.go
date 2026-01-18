package providers

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3UploadProvider struct {
	client   *s3.Client
	uploader *manager.Uploader
	bucket   string
	endpoint string
}

// S3Config holds the configuration for S3 uploads
type S3Config struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
}

func NewS3UploadProvider(cfg S3Config) (*S3UploadProvider, error) {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with custom endpoint if provided
	var client *s3.Client
	if cfg.Endpoint != "" {
		client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true // Required for LocalStack and some S3-compatible services
		})
	} else {
		client = s3.NewFromConfig(awsCfg)
	}

	// Create uploader using s3 manager
	uploader := manager.NewUploader(client)

	return &S3UploadProvider{
		client:   client,
		uploader: uploader,
		bucket:   cfg.Bucket,
		endpoint: cfg.Endpoint,
	}, nil
}

func (s *S3UploadProvider) UploadFile(file *multipart.FileHeader, path string) (string, error) {
	// Open source file
	source, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = source.Close() }()

	// Clean the path
	key := filepath.ToSlash(path)

	// Upload to S3 using the manager
	result, err := s.uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        source,
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Return the S3 URL from the upload result
	return result.Location, nil
}

func (s *S3UploadProvider) DeleteFile(path string) error {
	key := filepath.ToSlash(path)

	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}
