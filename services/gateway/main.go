// Package main is the entry point for the API gateway service.
//
//	@title			Ad-Bidding Platform API Gateway
//	@version		1.0
//	@description	HTTP → gRPC proxy that exposes Campaign, Bidder, and Analytics services.
//	@host			localhost:8080
//	@BasePath		/
//	@schemes		http
//
//go:generate swag init --generalInfo main.go --dir .,../../api-gateway/handler --output ../../api-gateway/docs
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Gvinay90/ad-bidding-platform/api-gateway/client"
	_ "github.com/Gvinay90/ad-bidding-platform/api-gateway/docs" // Swagger generated docs
	"github.com/Gvinay90/ad-bidding-platform/api-gateway/handler"
	"github.com/Gvinay90/ad-bidding-platform/api-gateway/router"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/config"
	"github.com/Gvinay90/ad-bidding-platform/internal/platform/logx"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config/local.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	log := logx.New(os.Stdout, cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(log)

	clients, err := client.New(
		cfg.Server.CampaignGRPCAddr,
		cfg.Server.BidderGRPCAddr,
		cfg.Server.AnalyticsGRPCAddr,
	)
	if err != nil {
		log.Error("failed to create gRPC clients", "err", err)
		os.Exit(1)
	}
	defer clients.Close()

	h := handler.New(clients)
	r := router.New(h, log)

	addr := fmt.Sprintf(":%d", cfg.Server.GatewayHTTPPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("api-gateway listening", "addr", addr, "swagger", addr+"/swagger/")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down api-gateway")
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error("shutdown error", "err", err)
	}
}
