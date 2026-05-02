package service

import (
	"context"

	"github.com/Gvinay90/ad-bidding-platform/internal/bidder/cache"
)

type Decision struct {
	HasBid            bool
	WinningCampaignID string
	PriceCents        int64
}

type BidService struct{ cache *cache.Cache }

func New(c *cache.Cache) *BidService { return &BidService{cache: c} }

func (s *BidService) Evaluate(ctx context.Context, t cache.Targeting) (Decision, error) {
	cands, err := s.cache.FindCandidates(ctx, t)
	if err != nil {
		return Decision{}, err
	}
	if len(cands) == 0 {
		return Decision{}, nil
	}

	winner := cands[0]
	for _, c := range cands[1:] {
		if c.BidPriceCents > winner.BidPriceCents {
			winner = c
		}
	}
	return Decision{HasBid: true, WinningCampaignID: winner.ID, PriceCents: winner.BidPriceCents}, nil
}
