package observe

import (
	"net/http"

	"log/slog"

	"github.com/dealer/dealer/pkg/errorreport"
	"github.com/dealer/dealer/pkg/grpcauth"
	"github.com/dealer/dealer/pkg/grpcmw"
	"github.com/dealer/dealer/pkg/health"
	"github.com/dealer/dealer/pkg/httplog"
	"github.com/dealer/dealer/pkg/metrics"
	"github.com/dealer/dealer/pkg/obslog"
	"google.golang.org/grpc"
)

// Init — структурированные логи сервиса.
func Init(service string) *slog.Logger {
	return obslog.Init(service)
}

// RegisterHTTP — /healthz, /readyz, /metrics.
func RegisterHTTP(mux *http.ServeMux, ready health.Check) {
	health.Register(mux, ready)
	metrics.RegisterHTTP(mux)
}

// WrapHTTP — логи, метрики и опциональная отправка HTTP 5xx в Kafka.
func WrapHTTP(service string, mux http.Handler, logger *slog.Logger) http.Handler {
	reporter := errorreport.NewFromEnv(service, logger)
	return httplog.Wrap(service, mux, logger, reporter)
}

// GRPCServerOptions — interceptors для grpc.NewServer. auth != nil включает JWT/RBAC на gRPC.
func GRPCServerOptions(service string, logger *slog.Logger, auth *grpcauth.Config) []grpc.ServerOption {
	reporter := errorreport.NewFromEnv(service, logger)
	chain := []grpc.UnaryServerInterceptor{
		grpcmw.UnaryServerInterceptor(service, logger, reporter),
	}
	if auth != nil && auth.JWTSecret != "" {
		chain = append(chain, grpcauth.UnaryServerInterceptor(*auth))
	}
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(chain...),
	}
}
