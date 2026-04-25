package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Gvinay90/ad-bidding-platform/internal/campaign/domain"
	"gorm.io/gorm"
)

var ErrCampaignNotFound = errors.New("campaign not found")

type campaignRow struct {
	ID            string `gorm:"primaryKey;size:36"`
	AdvertiserID  string `gorm:"not null;size:36"`
	Name          string `gorm:"not null;size:255"`
	BudgetCents   int64
	BidPriceCents int64
	Geo           string    `gorm:"index;size:255"`
	Device        string    `gorm:"index;size:32"`
	Category      string    `gorm:"index;size:64"`
	Status        string    `gorm:"index;size:16"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (campaignRow) TableName() string { return "campaigns" }

type CampaignRepository interface {
	CreateCampaign(ctx context.Context, campaign *domain.Campaign) error
	GetCampaignByID(ctx context.Context, id string) (*domain.Campaign, error)
	UpdateCampaign(ctx context.Context, campaign *domain.Campaign) error
	DeleteCampaign(ctx context.Context, id string) error
	ListByAdvertiserID(ctx context.Context, advertiserID string) ([]*domain.Campaign, error)
}

type GormCampaignRepo struct{ db *gorm.DB }

func NewGormCampaignRepo(db *gorm.DB) (*GormCampaignRepo, error) {
	if err := db.AutoMigrate(&campaignRow{}); err != nil {
		return nil, err
	}
	return &GormCampaignRepo{db: db}, nil
}

func toDomain(r *campaignRow) *domain.Campaign {
	return &domain.Campaign{
		ID:            r.ID,
		AdvertiserID:  r.AdvertiserID,
		Name:          r.Name,
		BudgetCents:   r.BudgetCents,
		BidPriceCents: r.BidPriceCents,
		Geo:           r.Geo,
		Device:        r.Device,
		Category:      r.Category,
		Status:        domain.Status(r.Status),
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

func fromDomain(c *domain.Campaign) *campaignRow {
	return &campaignRow{
		ID:            c.ID,
		AdvertiserID:  c.AdvertiserID,
		Name:          c.Name,
		BudgetCents:   c.BudgetCents,
		BidPriceCents: c.BidPriceCents,
		Geo:           c.Geo,
		Device:        c.Device,
		Category:      c.Category,
		Status:        string(c.Status),
	}
}

func (r *GormCampaignRepo) CreateCampaign(ctx context.Context, campaign *domain.Campaign) error {
	return r.db.WithContext(ctx).Create(fromDomain(campaign)).Error
}

func (r *GormCampaignRepo) GetCampaignByID(ctx context.Context, id string) (*domain.Campaign, error) {
	var row campaignRow
	if err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCampaignNotFound
		}
		return nil, err
	}
	return toDomain(&row), nil
}

func (r *GormCampaignRepo) UpdateCampaign(ctx context.Context, campaign *domain.Campaign) error {
	res := r.db.WithContext(ctx).Model(&campaignRow{}).Where("id = ?", campaign.ID).Updates(fromDomain(campaign))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrCampaignNotFound
	}
	return nil
}

func (r *GormCampaignRepo) DeleteCampaign(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Model(&campaignRow{}).Where("id = ?", id).Delete(&campaignRow{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrCampaignNotFound
	}
	return nil
}

func (r *GormCampaignRepo) ListByAdvertiserID(ctx context.Context, advertiserID string) ([]*domain.Campaign, error) {
	var rows []campaignRow
	q := r.db.WithContext(ctx).Order("created_at desc")
	if advertiserID != "" {
		q = q.Where("advertiser_id = ?", advertiserID)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Campaign, 0, len(rows))
	for i := range rows {
		out = append(out, toDomain(&rows[i]))
	}
	return out, nil
}
