package domain

import "time"

// BidEvent is one auction outcome persisted for analytics (matches bidder SNS payload).
type BidEvent struct {
	ID         string
	CampaignID string
	Won        bool
	PriceCents int64
	Geo        string
	Device     string
	Category   string
	Timestamp  time.Time
}

// Stats is aggregated spend / counts for a campaign.
type Stats struct {
	CampaignID string
	Wins       int64
	Bids       int64
	SpendCents int64
}
