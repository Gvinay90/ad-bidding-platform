package handler

import (
	"context"
	"time"

	"github.com/Gvinay90/ad-bidding-platform/internal/bidder/cache"
	"github.com/Gvinay90/ad-bidding-platform/internal/bidder/events"
	"github.com/Gvinay90/ad-bidding-platform/internal/bidder/service"
	bidderpb "github.com/Gvinay90/ad-bidding-platform/proto/bidder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BidderHandler struct {
	bidderpb.UnimplementedBidderServiceServer
	service    *service.BidService
	bidPublish *events.BidPublisher
}

func NewBidderHandler(svc *service.BidService, pub *events.BidPublisher) *BidderHandler {
	return &BidderHandler{service: svc, bidPublish: pub}
}

func (h *BidderHandler) EvaluateBid(ctx context.Context, r *bidderpb.BidRequest) (*bidderpb.BidResponse, error) {
	d, err := h.service.Evaluate(ctx, cache.Targeting{
		Geo: r.GetGeo(), Device: r.GetDevice(), Category: r.GetCategory(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if h.bidPublish != nil {
		go func() {
			pctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			h.bidPublish.PublishBidEvent(pctx, r, d)
		}()
	}
	return &bidderpb.BidResponse{
		HasBid: d.HasBid, WinningCampaignId: d.WinningCampaignID, PriceCents: d.PriceCents,
	}, nil
}
