package awsx

import (
	"context"
	"fmt"
	"strings"

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

func (c *Client) SNS() *sns.Client {
	return c.sns
}

func (c *Client) SQS() *sqs.Client {
	return c.sqs
}

// SNSTopicARN returns cfg.SNSTopic when it is already an ARN; otherwise builds
// arn:aws:sns:{region}:{account}:{name} (LocalStack uses account 000000000000).
func SNSTopicARN(cfg config.AWSConfig) string {
	t := strings.TrimSpace(cfg.SNSTopic)
	if t == "" {
		return ""
	}
	if strings.HasPrefix(t, "arn:") {
		return t
	}
	region := strings.TrimSpace(cfg.Region)
	if region == "" {
		region = "us-east-1"
	}
	acct := strings.TrimSpace(cfg.AccountID)
	if acct == "" {
		acct = "000000000000"
	}
	return fmt.Sprintf("arn:aws:sns:%s:%s:%s", region, acct, t)
}

// TopicNameFromConfig returns the topic name for CreateTopic (last segment if cfg.SNSTopic is an ARN).
func TopicNameFromConfig(cfg config.AWSConfig) string {
	t := strings.TrimSpace(cfg.SNSTopic)
	if t == "" {
		return ""
	}
	if strings.HasPrefix(t, "arn:") {
		parts := strings.Split(t, ":")
		if len(parts) >= 6 {
			return parts[len(parts)-1]
		}
	}
	return t
}

// EnsureSNSTopic returns the topic ARN to use for Publish. With a custom endpoint
// (e.g. LocalStack), it creates the topic if missing so local runs do not depend
// on deploy/localstack-init alone. For real AWS (no endpoint), it returns the
// configured ARN without creating anything.
func EnsureSNSTopic(ctx context.Context, client *sns.Client, cfg config.AWSConfig) (string, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return SNSTopicARN(cfg), nil
	}
	name := TopicNameFromConfig(cfg)
	if name == "" {
		return "", fmt.Errorf("aws: empty sns topic name")
	}
	out, err := client.CreateTopic(ctx, &sns.CreateTopicInput{Name: aws.String(name)})
	if err != nil {
		return "", err
	}
	if out.TopicArn != nil && *out.TopicArn != "" {
		return *out.TopicArn, nil
	}
	return SNSTopicARN(cfg), nil
}

func New(ctx context.Context, cfg config.AWSConfig) (*Client, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}
	if cfg.Endpoint != "" {
		ak := strings.TrimSpace(cfg.AccessKeyID)
		sk := strings.TrimSpace(cfg.SecretAccessKey)
		if ak == "" {
			ak = "test"
		}
		if sk == "" {
			sk = "test"
		}
		opts = append(opts, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(ak, sk, "")))
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
