package awsx

import (
	"context"

	"github.com/Gvinay90/ad-bidding-platform/internal/platform/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Client struct {
	sns *sns.Client
	sqs *sqs.Client
}

func New(ctx context.Context, cfg config.AWSConfig) (*Client, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}
	if cfg.Endpoint != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")))
	}
	awsConfig, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	snsClient := sns.NewFromConfig(awsConfig, func(o *sns.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})
	sqsClient := sqs.NewFromConfig(awsConfig, func(o *sqs.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})
	return &Client{
		sns: snsClient,
		sqs: sqsClient,
	}, nil
}
