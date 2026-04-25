package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Gvinay90/ad-bidding-platform/internal/campaign/domain"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/google/uuid"
)

type EventType string

const (
	Created EventType = "campaign.created"
	Updated EventType = "campaign.updated"
	Deleted EventType = "campaign.deleted"
)

type CampaignChanged struct {
	EventID    string           `json:"event_id"`
	EventType  EventType        `json:"event_type"`
	OccurredAt time.Time        `json:"occurred_at"`
	CampaignID string           `json:"campaign_id"`
	Snapshot   *domain.Campaign `json:"snapshot,omitempty"`
}

type Publisher struct {
	sns      *sns.Client
	topicArn string
}

func NewPublisher(sns *sns.Client, topicArn string) *Publisher {
	return &Publisher{
		sns:      sns,
		topicArn: topicArn,
	}
}

func (p *Publisher) PublishCampaignChanged(ctx context.Context, t EventType, c *domain.Campaign, deletedID string) error {
	ev := CampaignChanged{
		EventID:    uuid.New().String(),
		EventType:  t,
		OccurredAt: time.Now(),
	}
	if c != nil {
		ev.CampaignID = c.ID
		ev.Snapshot = c
	} else {
		ev.CampaignID = deletedID
	}
	body, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	_, err = p.sns.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(p.topicArn),
		Message:  aws.String(string(body)),
	})
	return err
}
