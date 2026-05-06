package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/Gvinay90/ad-bidding-platform/internal/analytics/domain"
	"github.com/Gvinay90/ad-bidding-platform/internal/analytics/service"
	"github.com/google/uuid"
)

type snsEnvelope struct {
	Message string `json:"Message"`
}

type bidEventMsg struct {
	EventID           string    `json:"event_id"`
	OccurredAt        time.Time `json:"occurred_at"`
	RequestID         string    `json:"request_id"`
	WinningCampaignID string    `json:"winning_campaign_id"`
	PriceCents        int64     `json:"price_cents"`
	Targeting         struct {
		Geo, Device, Category string
	} `json:"targeting"`
}

type Consumer struct {
	sqs      *sqs.Client
	queueURL string
	svc      *service.AnalyticsService
}

func NewConsumer(c *sqs.Client, queueURL string, svc *service.AnalyticsService) *Consumer {
	return &Consumer{sqs: c, queueURL: queueURL, svc: svc}
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
			if ctx.Err() != nil {
				return
			}
			slog.Warn("analytics sqs receive", "err", err)
			time.Sleep(time.Second)
			continue
		}
		for _, m := range out.Messages {
			if err := c.handle(ctx, m.Body); err != nil {
				slog.Warn("analytics handle message", "err", err)
				continue
			}
			_, _ = c.sqs.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(c.queueURL),
				ReceiptHandle: m.ReceiptHandle,
			})
		}
	}
}

func (c *Consumer) handle(ctx context.Context, body *string) error {
	if body == nil {
		return nil
	}
	var env snsEnvelope
	if err := json.Unmarshal([]byte(*body), &env); err != nil {
		return err
	}
	inner := strings.TrimSpace(env.Message)
	if inner == "" {
		return nil
	}
	// CampaignChanged messages share this queue via SNS fan-out; skip them.
	var probe struct {
		EventType string `json:"event_type"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal([]byte(inner), &probe); err != nil {
		return err
	}
	if probe.EventType != "" && strings.HasPrefix(probe.EventType, "campaign.") {
		return nil
	}

	var msg bidEventMsg
	if err := json.Unmarshal([]byte(inner), &msg); err != nil {
		return err
	}
	ev := &domain.BidEvent{
		ID:         msg.EventID,
		CampaignID: msg.WinningCampaignID,
		Won:        msg.WinningCampaignID != "",
		PriceCents: msg.PriceCents,
		Geo:        msg.Targeting.Geo,
		Device:     msg.Targeting.Device,
		Category:   msg.Targeting.Category,
		Timestamp:  msg.OccurredAt,
	}
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now().UTC()
	}
	if ev.ID == "" {
		ev.ID = uuid.New().String()
	}
	return c.svc.RecordBid(ctx, ev)
}
