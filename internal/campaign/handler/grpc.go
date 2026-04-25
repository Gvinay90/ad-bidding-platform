package handler

import (
	"context"

	"github.com/Gvinay90/ad-bidding-platform/internal/campaign/service"
	campaignpb "github.com/Gvinay90/ad-bidding-platform/proto/campaign"
)

type CampaignHandler struct {
	campaignpb.UnimplementedCampaignServiceServer
	service *service.CampaignService
}

func NewCampaignHandler(service *service.CampaignService) *CampaignHandler {
	return &CampaignHandler{service: service}
}

func (h *CampaignHandler) CreateCampaign(ctx context.Context, req *campaignpb.CreateCampaignRequest) (*campaignpb.CampaignResponse, error) {
	res, err := h.service.CreateCampaign(ctx, req)
	if err != nil {
		return nil, err
	}
	return &campaignpb.CampaignResponse{Campaign: res.Campaign}, nil
}

func (h *CampaignHandler) GetCampaign(ctx context.Context, req *campaignpb.GetCampaignRequest) (*campaignpb.CampaignResponse, error) {
	res, err := h.service.GetCampaign(ctx, req)
	if err != nil {
		return nil, err
	}
	return &campaignpb.CampaignResponse{Campaign: res.Campaign}, nil
}

func (h *CampaignHandler) ListCampaigns(ctx context.Context, req *campaignpb.ListCampaignsRequest) (*campaignpb.ListCampaignsResponse, error) {
	campaigns, err := h.service.ListCampaigns(ctx, req.AdvertiserId)
	if err != nil {
		return nil, err
	}
	campaignsPb := make([]*campaignpb.Campaign, len(campaigns))
	for i, campaign := range campaigns {
		campaignsPb[i] = &campaignpb.Campaign{
			Id:            campaign.ID,
			AdvertiserId:  campaign.AdvertiserID,
			Name:          campaign.Name,
			BudgetCents:   campaign.BudgetCents,
			BidPriceCents: campaign.BidPriceCents,
			Geo:           campaign.Geo,
			Device:        campaign.Device,
			Category:      campaign.Category,
			Status:        string(campaign.Status),
		}
	}
	return &campaignpb.ListCampaignsResponse{Campaigns: campaignsPb}, nil
}

func (h *CampaignHandler) UpdateCampaign(ctx context.Context, req *campaignpb.UpdateCampaignRequest) (*campaignpb.CampaignResponse, error) {
	res, err := h.service.UpdateCampaign(ctx, req)
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

func (h *CampaignHandler) DeleteCampaign(ctx context.Context, req *campaignpb.DeleteCampaignRequest) (*campaignpb.DeleteCampaignResponse, error) {
	res, err := h.service.DeleteCampaign(ctx, req)
	if err != nil {
		return nil, err
	}
	return &campaignpb.DeleteCampaignResponse{Message: res.Message}, nil
}
