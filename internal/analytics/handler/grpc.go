package handler

import (
	"context"
	"errors"

	"github.com/Gvinay90/ad-bidding-platform/internal/analytics/service"
	analyticspb "github.com/Gvinay90/ad-bidding-platform/proto/analytics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AnalyticsHandler struct {
	analyticspb.UnimplementedAnalyticsServiceServer
	svc *service.AnalyticsService
}

func NewAnalyticsHandler(svc *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

func (h *AnalyticsHandler) GetCampaignStats(ctx context.Context, req *analyticspb.CampaignStatsRequest) (*analyticspb.CampaignStatsResponse, error) {
	stats, err := h.svc.GetStats(ctx, req.GetCampaignId())
	if err != nil {
		if errors.Is(err, service.ErrInvalidID) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &analyticspb.CampaignStatsResponse{
		CampaignId: stats.CampaignID,
		Wins:       stats.Wins,
		Bids:       stats.Bids,
		SpendCents: stats.SpendCents,
	}, nil
}
