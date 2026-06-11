package grpcmw

import (
	"context"
	"log/slog"
	"time"

	"github.com/dealer/dealer/pkg/errorevent"
	"github.com/dealer/dealer/pkg/errorreport"
	"github.com/dealer/dealer/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor — логи, метрики и опциональная отправка server-side ошибок в Kafka.
func UnaryServerInterceptor(service string, logger *slog.Logger, reporter *errorreport.Reporter) grpc.UnaryServerInterceptor {
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
			if reporter != nil && shouldReportGRPC(code) {
				ev := errorevent.New(service, "server_error", severityForGRPC(code), err.Error())
				ev.GRPCCode = code.String()
				ev.Route = info.FullMethod
				reporter.Report(ev)
			}
		}
		metrics.RecordGRPC(service, info.FullMethod, code.String(), duration)
		return resp, err
	}
}

func shouldReportGRPC(code codes.Code) bool {
	switch code {
	case codes.Internal, codes.Unknown, codes.Unavailable, codes.DataLoss, codes.Aborted:
		return true
	default:
		return false
	}
}

func severityForGRPC(code codes.Code) string {
	if code == codes.Internal || code == codes.Unknown || code == codes.DataLoss {
		return "error"
	}
	return "warn"
}
