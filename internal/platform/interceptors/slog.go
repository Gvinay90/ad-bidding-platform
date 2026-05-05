package interceptors

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
)

// UnarySlogInterceptor logs each unary RPC with optional service name in attributes.
func UnarySlogInterceptor(l *slog.Logger, service string) grpc.UnaryServerInterceptor {
	if l == nil {
		l = slog.Default()
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (any, error) {
		start := time.Now()
		if service != "" {
			l.InfoContext(ctx, "grpc request", "service", service, "method", info.FullMethod)
		} else {
			l.InfoContext(ctx, "grpc request", "method", info.FullMethod)
		}
		resp, err := h(ctx, req)
		ms := time.Since(start).Milliseconds()
		if err != nil {
			if service != "" {
				l.ErrorContext(ctx, "grpc request failed", "service", service, "method", info.FullMethod, "duration_ms", ms, "err", err)
			} else {
				l.ErrorContext(ctx, "grpc request failed", "method", info.FullMethod, "duration_ms", ms, "err", err)
			}
			return resp, err
		}
		if service != "" {
			l.InfoContext(ctx, "grpc request completed", "service", service, "method", info.FullMethod, "duration_ms", ms)
		} else {
			l.InfoContext(ctx, "grpc request completed", "method", info.FullMethod, "duration_ms", ms)
		}
		return resp, err
	}
}
