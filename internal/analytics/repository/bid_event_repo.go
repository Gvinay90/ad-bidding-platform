package repository

import (
	"context"
	"time"

	"github.com/Gvinay90/ad-bidding-platform/internal/analytics/domain"
	"gorm.io/gorm"
)

type bidEventRow struct {
	ID         string `gorm:"primaryKey;size:36"`
	CampaignID string `gorm:"index;size:36"`
	Won        bool   `gorm:"index"`
	PriceCents int64
	Geo        string    `gorm:"size:64"`
	Device     string    `gorm:"size:32"`
	Category   string    `gorm:"size:64"`
	Timestamp  time.Time `gorm:"index"`
}

func (bidEventRow) TableName() string { return "bid_events" }

type BidEventRepository interface {
	Insert(ctx context.Context, ev *domain.BidEvent) error
	GetStats(ctx context.Context, campaignID string) (*domain.Stats, error)
}

type GormBidEventRepo struct{ db *gorm.DB }

func NewGormBidEventRepo(db *gorm.DB) (*GormBidEventRepo, error) {
	if err := db.AutoMigrate(&bidEventRow{}); err != nil {
		return nil, err
	}
	return &GormBidEventRepo{db: db}, nil
}

func (r *GormBidEventRepo) Insert(ctx context.Context, ev *domain.BidEvent) error {
	row := &bidEventRow{
		ID: ev.ID, CampaignID: ev.CampaignID, Won: ev.Won,
		PriceCents: ev.PriceCents, Geo: ev.Geo, Device: ev.Device,
		Category: ev.Category, Timestamp: ev.Timestamp,
	}
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *GormBidEventRepo) GetStats(ctx context.Context, campaignID string) (*domain.Stats, error) {
	var agg struct {
		Wins       int64
		Bids       int64
		SpendCents int64
	}
	err := r.db.WithContext(ctx).Raw(`
SELECT
  SUM(CASE WHEN won THEN 1 ELSE 0 END) AS wins,
  COUNT(*) AS bids,
  SUM(CASE WHEN won THEN price_cents ELSE 0 END) AS spend_cents
FROM bid_events WHERE campaign_id = ?`, campaignID).Scan(&agg).Error
	if err != nil {
		return nil, err
	}
	return &domain.Stats{
		CampaignID: campaignID,
		Wins:       agg.Wins,
		Bids:       agg.Bids,
		SpendCents: agg.SpendCents,
	}, nil
}
