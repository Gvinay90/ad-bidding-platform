package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Gvinay90/ad-bidding-platform/internal/analytics/events"
	"github.com/Gvinay90/ad-bidding-platform/internal/analytics/handler"
	"github.com/Gvinay90/ad-bidding-platform/internal/analytics/repository"
	"github.com/Gvinay90/ad-bidding-platform/internal/analytics/service"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/awsx"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/config"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/db"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/interceptors"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/logx"
	analyticspb "github.com/Gvinay90/ad-bidding-platform/proto/analytics"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.Open(&cfg.Database)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	repo, err := repository.NewGormBidEventRepo(database)
	if err != nil {
		slog.Error("failed to init bid_events repository", "err", err)
		os.Exit(1)
	}
	svc := service.New(repo)

	awsClient, err := awsx.New(ctx, cfg.AWS)
	if err != nil {
		slog.Error("failed to create AWS client", "err", err)
		os.Exit(1)
	}
	queueURL, err := awsx.QueueURL(ctx, awsClient.SQS(), cfg.AWS, cfg.AWS.AnalyticsQueue)
	if err != nil {
		slog.Error("failed to resolve analytics SQS queue URL", "queue", cfg.AWS.AnalyticsQueue, "err", err)
		os.Exit(1)
	}
	slog.Info("analytics sqs consumer queue", "queue_url", queueURL)
	cons := events.NewConsumer(awsClient.SQS(), queueURL, svc)
	go cons.Run(ctx)

	h := handler.NewAnalyticsHandler(svc)
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.UnarySlogInterceptor(logger, "analytics")),
	)
	analyticspb.RegisterAnalyticsServiceServer(grpcServer, h)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.AnalyticsGRPCPort))
	if err != nil {
		slog.Error("failed to listen", "err", err)
		os.Exit(1)
	}
	slog.Info("analytics service listening", "addr", lis.Addr().String(), "reflection", true)

	errCh := make(chan error, 1)
	go func() {
		errCh <- grpcServer.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutting down analytics service")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		stopped := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
			slog.Info("grpc server stopped")
		case <-shutdownCtx.Done():
			grpcServer.Stop()
			slog.Warn("grpc server forced stop")
		}
	case err := <-errCh:
		if err != nil {
			slog.Error("grpc server error", "err", err)
			os.Exit(1)
		}
	}
}
