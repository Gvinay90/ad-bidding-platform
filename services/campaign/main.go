package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/Gvinay90/ad-bidding-platform/internal/campaign/events"
	"github.com/Gvinay90/ad-bidding-platform/internal/campaign/handler"
	"github.com/Gvinay90/ad-bidding-platform/internal/campaign/repository"
	"github.com/Gvinay90/ad-bidding-platform/internal/campaign/service"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/awsx"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/config"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/db"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/logx"
	campaignpb "github.com/Gvinay90/ad-bidding-platform/proto/campaign"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load("config/local.yaml")
	if err != nil {
		slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})).
			Error("failed to load config", "err", err)
		os.Exit(1)
	}
	logger := logx.New(os.Stdout, cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(logger)

	slog.Info("campaign service starting")
	ctx := context.Background()

	database, err := db.Open(&cfg.Database)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	repo, err := repository.NewGormCampaignRepo(database)
	if err != nil {
		slog.Error("failed to create campaign repository", "err", err)
		os.Exit(1)
	}
	awsClient, err := awsx.New(ctx, cfg.AWS)
	if err != nil {
		slog.Error("failed to create AWS client", "err", err)
		os.Exit(1)
	}
	topicARN, err := awsx.EnsureSNSTopic(ctx, awsClient.SNS(), cfg.AWS)
	if err != nil {
		slog.Error("failed to ensure SNS topic", "err", err)
		os.Exit(1)
	}
	slog.Info("sns topic ready", "topic_arn", topicARN)

	publisher := events.NewPublisher(awsClient.SNS(), topicARN)
	svc := service.NewCampaignService(repo, publisher)
	h := handler.NewCampaignHandler(svc)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(handler.UnarySlogInterceptor(logger)),
	)
	campaignpb.RegisterCampaignServiceServer(grpcServer, h)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.CampaignGRPCPort))
	if err != nil {
		slog.Error("failed to listen", "err", err)
		os.Exit(1)
	}
	slog.Info("campaign service listening", "addr", lis.Addr().String(), "reflection", true)
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("grpc server stopped with error", "err", err)
		os.Exit(1)
	}
}
