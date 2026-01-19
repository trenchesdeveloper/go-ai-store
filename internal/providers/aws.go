package providers

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// CreateAWSConfig creates an AWS configuration
// For custom endpoints (e.g., MinIO, LocalStack), uses config.WithBaseEndpoint
func CreateAWSConfig(ctx context.Context, endpoint, region string) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	if endpoint != "" {
		opts = append(opts, config.WithBaseEndpoint(endpoint))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, err
	}

	return cfg, nil
}
