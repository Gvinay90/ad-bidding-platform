package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Gvinay90/ad-bidding-platform/internal/campaign/domain"
	"github.com/Gvinay90/ad-bidding-platform/internal/campaign/events"
	"github.com/Gvinay90/ad-bidding-platform/internal/campaign/repository"
	campaignpb "github.com/Gvinay90/ad-bidding-platform/proto/campaign"
	"github.com/google/uuid"
)

var (
	ErrInvalidID = errors.New("invalid ID")
	ErrInvalid   = errors.New("invalid request")
)

type CampaignService struct {
	repo      repository.CampaignRepository
	publisher *events.Publisher
}

func NewCampaignService(repo repository.CampaignRepository, publisher *events.Publisher) *CampaignService {
	return &CampaignService{repo: repo, publisher: publisher}
}

func (s *CampaignService) CreateCampaign(ctx context.Context, req *campaignpb.CreateCampaignRequest) (*campaignpb.CampaignResponse, error) {
	if strings.TrimSpace(req.Name) == "" || req.BudgetCents <= 0 || req.BidPriceCents <= 0 {
		return nil, ErrInvalid
	}
	c := &domain.Campaign{
		ID:            uuid.New().String(),
		AdvertiserID:  req.AdvertiserId,
		Name:          req.Name,
		BudgetCents:   req.BudgetCents,
		BidPriceCents: req.BidPriceCents,
		Geo:           req.Geo,
		Device:        req.Device,
		Category:      req.Category,
		Status:        domain.StatusActive,
	}
	if err := s.repo.CreateCampaign(ctx, c); err != nil {
		return nil, err
	}
	if err := s.publisher.PublishCampaignChanged(ctx, events.Created, c, ""); err != nil {
		return nil, err
	}
	return &campaignpb.CampaignResponse{Campaign: &campaignpb.Campaign{
		Id:            c.ID,
		AdvertiserId:  c.AdvertiserID,
		Name:          c.Name,
		BudgetCents:   c.BudgetCents,
		BidPriceCents: c.BidPriceCents,
		Geo:           c.Geo,
		Device:        c.Device,
		Category:      c.Category,
		Status:        string(c.Status),
	}}, nil
}

func (s *CampaignService) GetCampaign(ctx context.Context, req *campaignpb.GetCampaignRequest) (*campaignpb.CampaignResponse, error) {
	if strings.TrimSpace(req.Id) == "" {
		return nil, ErrInvalidID
	}
	res, err := s.repo.GetCampaignByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &campaignpb.CampaignResponse{Campaign: &campaignpb.Campaign{
		Id:            res.ID,
		AdvertiserId:  res.AdvertiserID,
		Name:          res.Name,
		BudgetCents:   res.BudgetCents,
		BidPriceCents: res.BidPriceCents,
		Geo:           res.Geo,
		Device:        res.Device,
		Category:      res.Category,
		Status:        string(res.Status),
	}}, nil
}

func (s *CampaignService) ListCampaigns(ctx context.Context, advertiserID string) ([]*domain.Campaign, error) {
	res, err := s.repo.ListByAdvertiserID(ctx, advertiserID)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *CampaignService) UpdateCampaign(ctx context.Context, req *campaignpb.UpdateCampaignRequest) (*domain.Campaign, error) {
	if strings.TrimSpace(req.Id) == "" || req.BudgetCents <= 0 || req.BidPriceCents <= 0 {
		return nil, ErrInvalid
	}
	c, err := s.repo.GetCampaignByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	c.Name = req.Name
	c.BudgetCents = req.BudgetCents
	c.BidPriceCents = req.BidPriceCents
	c.Geo = req.Geo
	c.Device = req.Device
	c.Category = req.Category
	c.Status = domain.Status(req.Status)
	if err := s.repo.UpdateCampaign(ctx, c); err != nil {
		return nil, err
	}
	if err := s.publisher.PublishCampaignChanged(ctx, events.Updated, c, ""); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CampaignService) DeleteCampaign(ctx context.Context, req *campaignpb.DeleteCampaignRequest) (*campaignpb.DeleteCampaignResponse, error) {
	if strings.TrimSpace(req.Id) == "" {
		return nil, ErrInvalidID
	}
	err := s.repo.DeleteCampaign(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &campaignpb.DeleteCampaignResponse{Message: "Campaign deleted successfully"}, nil
}
