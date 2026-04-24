package redisx

import (
	"context"
	"time"

	"github.com/Gvinay90/ad-bidding-platform/internal/platform/config"
	"github.com/redis/go-redis/v9"
)

func New(cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
