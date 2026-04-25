package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/Gvinay90/ad-bidding-platform/internal/campaign/events"
	"github.com/Gvinay90/ad-bidding-platform/internal/campaign/handler"
	"github.com/Gvinay90/ad-bidding-platform/internal/campaign/repository"
	"github.com/Gvinay90/ad-bidding-platform/internal/campaign/service"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/awsx"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/config"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/db"
	campaignpb "github.com/Gvinay90/ad-bidding-platform/proto/campaign"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Println("campaign service: starting")
	cfg, err := config.Load("config/local.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	database, err := db.Open(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	repo, err := repository.NewGormCampaignRepo(database)
	if err != nil {
		log.Fatalf("Failed to create campaign repository: %v", err)
	}
	awsClient, err := awsx.New(context.Background(), cfg.AWS)
	if err != nil {
		log.Fatalf("Failed to create AWS client: %v", err)
	}
	topicARN, err := awsx.EnsureSNSTopic(context.Background(), awsClient.SNS(), cfg.AWS)
	if err != nil {
		log.Fatalf("Failed to ensure SNS topic: %v", err)
	}
	publisher := events.NewPublisher(awsClient.SNS(), topicARN)
	service := service.NewCampaignService(repo, publisher)
	handler := handler.NewCampaignHandler(service)
	grpcServer := grpc.NewServer()
	campaignpb.RegisterCampaignServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.CampaignGRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("campaign service: listening on %s (gRPC + reflection)\n", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
