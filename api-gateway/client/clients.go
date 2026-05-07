package client

import (
	analyticspb "github.com/Gvinay90/ad-bidding-platform/proto/analytics"
	bidderpb "github.com/Gvinay90/ad-bidding-platform/proto/bidder"
	campaignpb "github.com/Gvinay90/ad-bidding-platform/proto/campaign"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Clients holds gRPC clients for every downstream service.
type Clients struct {
	Campaign  campaignpb.CampaignServiceClient
	Bidder    bidderpb.BidderServiceClient
	Analytics analyticspb.AnalyticsServiceClient
	conns     []*grpc.ClientConn
}

func New(campaignAddr, bidderAddr, analyticsAddr string) (*Clients, error) {
	dial := func(addr string) (*grpc.ClientConn, error) {
		return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := dial(campaignAddr)
	if err != nil {
		return nil, err
	}
	bc, err := dial(bidderAddr)
	if err != nil {
		return nil, err
	}
	ac, err := dial(analyticsAddr)
	if err != nil {
		return nil, err
	}
	return &Clients{
		Campaign:  campaignpb.NewCampaignServiceClient(cc),
		Bidder:    bidderpb.NewBidderServiceClient(bc),
		Analytics: analyticspb.NewAnalyticsServiceClient(ac),
		conns:     []*grpc.ClientConn{cc, bc, ac},
	}, nil
}

func (c *Clients) Close() {
	for _, x := range c.conns {
		_ = x.Close()
	}
}
