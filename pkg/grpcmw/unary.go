package grpcmw

import (
	"context"
	"log/slog"
	"time"

	"github.com/dealer/dealer/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor — логи и метрики unary RPC.
func UnaryServerInterceptor(service string, logger *slog.Logger) grpc.UnaryServerInterceptor {
	if logger == nil {
		logger = slog.Default()
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err)
		duration := time.Since(start)

		logger.Info("grpc request",
			"method", info.FullMethod,
			"code", code.String(),
			"duration_ms", duration.Milliseconds(),
		)
		if err != nil && code != codes.OK {
			logger.Warn("grpc request failed",
				"method", info.FullMethod,
				"code", code.String(),
				"err", err.Error(),
			)
		}
		metrics.RecordGRPC(service, info.FullMethod, code.String(), duration)
		return resp, err
	}
}
