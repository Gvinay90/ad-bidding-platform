package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Gvinay90/ad-bidding-platform/internal/bidder/service"
	bidderpb "github.com/Gvinay90/ad-bidding-platform/proto/bidder"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/google/uuid"
)

// BidPublisher sends BidEvent payloads to SNS for the analytics service (same topic pattern as the lab).
type BidPublisher struct {
	sns      *sns.Client
	topicArn string
}

func NewBidPublisher(sns *sns.Client, topicArn string) *BidPublisher {
	return &BidPublisher{sns: sns, topicArn: topicArn}
}

type bidEventWire struct {
	EventID           string    `json:"event_id"`
	OccurredAt        time.Time `json:"occurred_at"`
	RequestID         string    `json:"request_id"`
	WinningCampaignID string    `json:"winning_campaign_id"`
	PriceCents        int64     `json:"price_cents"`
	Targeting         struct {
		Geo      string `json:"geo"`
		Device   string `json:"device"`
		Category string `json:"category"`
	} `json:"targeting"`
}

// PublishBidEvent publishes asynchronously; errors are ignored on the hot path (lab fire-and-forget).
func (p *BidPublisher) PublishBidEvent(ctx context.Context, req *bidderpb.BidRequest, d service.Decision) {
	if p == nil || p.sns == nil || p.topicArn == "" {
		return
	}
	ev := bidEventWire{
		EventID:           uuid.New().String(),
		OccurredAt:        time.Now().UTC(),
		RequestID:         req.GetRequestId(),
		WinningCampaignID: d.WinningCampaignID,
		PriceCents:        d.PriceCents,
	}
	ev.Targeting.Geo = req.GetGeo()
	ev.Targeting.Device = req.GetDevice()
	ev.Targeting.Category = req.GetCategory()
	body, err := json.Marshal(ev)
	if err != nil {
		return
	}
	_, _ = p.sns.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(p.topicArn),
		Message:  aws.String(string(body)),
	})
}
