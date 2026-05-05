package events

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/Gvinay90/ad-bidding-platform/internal/bidder/cache"
)

// snsEnvelope is the body SNS writes to SQS when raw delivery is off.
type snsEnvelope struct {
	Message string `json:"Message"`
}

type campaignSnapshot struct {
	ID                            string `json:"ID"`
	BidPriceCents                 int64  `json:"BidPriceCents"`
	Geo, Device, Category, Status string
}
type campaignChanged struct {
	EventType  string            `json:"event_type"`
	CampaignID string            `json:"campaign_id"`
	Snapshot   *campaignSnapshot `json:"snapshot"`
}

type Consumer struct {
	sqs      *sqs.Client
	queueURL string
	cache    *cache.Cache
}

func NewConsumer(c *sqs.Client, queueURL string, ca *cache.Cache) *Consumer {
	return &Consumer{sqs: c, queueURL: queueURL, cache: ca}
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		out, err := c.sqs.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(c.queueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     10,
		})
		if err != nil {
			log.Printf("sqs receive: %v", err)
			time.Sleep(time.Second)
			continue
		}
		for _, m := range out.Messages {
			c.handle(ctx, m.Body)
			_, _ = c.sqs.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl: aws.String(c.queueURL), ReceiptHandle: m.ReceiptHandle,
			})
		}
	}
}

func (c *Consumer) handle(ctx context.Context, body *string) {
	if body == nil {
		return
	}
	var env snsEnvelope
	if err := json.Unmarshal([]byte(*body), &env); err != nil {
		return
	}
	var ev campaignChanged
	if err := json.Unmarshal([]byte(env.Message), &ev); err != nil {
		return
	}
	switch ev.EventType {
	case "campaign.deleted":
		_ = c.cache.Delete(ctx, ev.CampaignID)
	default:
		if ev.Snapshot == nil {
			return
		}
		_ = c.cache.Upsert(ctx, ev.Snapshot.ID, ev.Snapshot.BidPriceCents,
			ev.Snapshot.Geo, ev.Snapshot.Device, ev.Snapshot.Category, ev.Snapshot.Status)
	}
}
