package handler

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
)

// UnarySlogInterceptor logs each unary RPC with method, duration, and errors.
func UnarySlogInterceptor(l *slog.Logger) grpc.UnaryServerInterceptor {
	if l == nil {
		l = slog.Default()
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (any, error) {
		start := time.Now()
		l.InfoContext(ctx, "grpc request", "method", info.FullMethod)
		resp, err := h(ctx, req)
		ms := time.Since(start).Milliseconds()
		if err != nil {
			l.ErrorContext(ctx, "grpc request failed", "method", info.FullMethod, "duration_ms", ms, "err", err)
			return resp, err
		}
		l.InfoContext(ctx, "grpc request completed", "method", info.FullMethod, "duration_ms", ms)
		return resp, err
	}
}
