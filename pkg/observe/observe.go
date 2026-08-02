package observe

import (
	"net/http"
	"os"
	"strings"

	"log/slog"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/dealer/dealer/pkg/errorreport"
	"github.com/dealer/dealer/pkg/grpcauth"
	"github.com/dealer/dealer/pkg/grpcmw"
	"github.com/dealer/dealer/pkg/health"
	"github.com/dealer/dealer/pkg/httplog"
	"github.com/dealer/dealer/pkg/metrics"
	"github.com/dealer/dealer/pkg/obsenv"
	"github.com/dealer/dealer/pkg/obslog"
	"github.com/dealer/dealer/pkg/obspath"
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

// WrapHTTP — логи, метрики, трейсинг и опциональная отправка HTTP 5xx в Kafka.
func WrapHTTP(service string, mux http.Handler, logger *slog.Logger) http.Handler {
	reporter := errorreport.NewFromEnv(service, logger)
	handler := httplog.Wrap(service, mux, logger, reporter)
	if !tracingEnabled() {
		return handler
	}
	return otelhttp.NewHandler(handler, service,
		otelhttp.WithSpanNameFormatter(spanNameFormatter),
		otelhttp.WithFilter(func(r *http.Request) bool {
			return !obspath.IsProbe(r.URL.Path)
		}),
	)
}

func spanNameFormatter(_ string, r *http.Request) string {
	return r.Method + " " + r.URL.Path
}

func tracingEnabled() bool {
	if os.Getenv(obsenv.OtlpEndpoint) == "" ||
		strings.EqualFold(os.Getenv(obsenv.TracesExporter), "none") {
		return false
	}
	return true
}

// GRPCServerOptions — interceptors для grpc.NewServer. auth != nil включает JWT/RBAC на gRPC.
func GRPCServerOptions(service string, logger *slog.Logger, auth *grpcauth.Config) []grpc.ServerOption {
	reporter := errorreport.NewFromEnv(service, logger)
	chain := []grpc.UnaryServerInterceptor{
		otelgrpc.UnaryServerInterceptor(),
		grpcmw.UnaryServerInterceptor(service, logger, reporter),
	}
	if auth != nil && auth.JWTSecret != "" {
		chain = append(chain, grpcauth.UnaryServerInterceptor(*auth))
	}
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(chain...),
	}
}
