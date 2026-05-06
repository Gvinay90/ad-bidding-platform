package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Gvinay90/ad-bidding-platform/internal/analytics/domain"
	"github.com/Gvinay90/ad-bidding-platform/internal/analytics/repository"
)

var ErrInvalidID = errors.New("campaign_id required")

type AnalyticsService struct {
	repo repository.BidEventRepository
}

func New(repo repository.BidEventRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

// RecordBid persists a bid outcome from the SQS consumer.
func (s *AnalyticsService) RecordBid(ctx context.Context, ev *domain.BidEvent) error {
	if ev == nil || strings.TrimSpace(ev.CampaignID) == "" {
		return nil
	}
	return s.repo.Insert(ctx, ev)
}

func (s *AnalyticsService) GetStats(ctx context.Context, campaignID string) (*domain.Stats, error) {
	if strings.TrimSpace(campaignID) == "" {
		return nil, ErrInvalidID
	}
	return s.repo.GetStats(ctx, campaignID)
}
