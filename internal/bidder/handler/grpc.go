package handler

import (
	"context"

	"github.com/Gvinay90/ad-bidding-platform/internal/bidder/cache"
	"github.com/Gvinay90/ad-bidding-platform/internal/bidder/service"
	bidderpb "github.com/Gvinay90/ad-bidding-platform/proto/bidder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BidderHandler struct {
	bidderpb.UnimplementedBidderServiceServer
	service *service.BidService
}

func NewBidderHandler(service *service.BidService) *BidderHandler {
	return &BidderHandler{service: service}
}
func (h *BidderHandler) EvaluateBid(ctx context.Context, r *bidderpb.BidRequest) (*bidderpb.BidResponse, error) {
	d, err := h.service.Evaluate(ctx, cache.Targeting{
		Geo: r.GetGeo(), Device: r.GetDevice(), Category: r.GetCategory(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &bidderpb.BidResponse{
		HasBid: d.HasBid, WinningCampaignId: d.WinningCampaignID, PriceCents: d.PriceCents,
	}, nil
}
