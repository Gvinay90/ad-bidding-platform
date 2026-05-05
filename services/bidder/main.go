package bidder

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/Gvinay90/ad-bidding-platform/internal/bidder/cache"
	"github.com/Gvinay90/ad-bidding-platform/internal/bidder/events"
	"github.com/Gvinay90/ad-bidding-platform/internal/bidder/handler"
	"github.com/Gvinay90/ad-bidding-platform/internal/bidder/service"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/awsx"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/config"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/interceptors"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/logx"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/redisx"
	bidderpb "github.com/Gvinay90/ad-bidding-platform/proto/bidder"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load("config/local.yaml")
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}
	logger := logx.New(os.Stdout, cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(logger)

	ctx := context.Background()

	rdb, err := redisx.New(&cfg.Redis)
	if err != nil {
		slog.Error("failed to connect to redis", "err", err)
		os.Exit(1)
	}
	c := cache.NewCache(rdb)
	aws, _ := awsx.New(ctx, cfg.AWS)
	queueURL := fmt.Sprintf("%s/000000000000/%s", cfg.AWS.Endpoint, cfg.AWS.BidderQueue)
	cons := events.NewConsumer(aws.SQS(), queueURL, c)
	go cons.Run(ctx) // run consumer in background

	svc := service.New(c)
	hnd := handler.NewBidderHandler(svc)
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.UnarySlogInterceptor(logger, "bidder")),
	)
	bidderpb.RegisterBidderServiceServer(grpcServer, hnd)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.BidderGRPCPort))
	if err != nil {
		slog.Error("failed to listen", "err", err)
		os.Exit(1)
	}
	slog.Info("bidder service listening", "addr", lis.Addr().String(), "reflection", true)
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("grpc server stopped with error", "err", err)
		os.Exit(1)
	}
}
