package metrics

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/dealer/dealer/pkg/obsenv"
	"github.com/dealer/dealer/pkg/obspath"
)

// Имена метрик Prometheus (для регистрации и тестов).
const (
	HTTPRequestsTotalName        = "http_requests_total"
	HTTPRequestDurationName      = "http_request_duration_seconds"
	GRPCRequestsTotalName        = "grpc_requests_total"
	GRPCRequestDurationName      = "grpc_request_duration_seconds"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: HTTPRequestsTotalName,
		Help: "Total HTTP requests",
	}, []string{"service", "method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    HTTPRequestDurationName,
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "method", "path"})

	grpcRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: GRPCRequestsTotalName,
		Help: "Total gRPC unary requests",
	}, []string{"service", "method", "code"})

	grpcRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    GRPCRequestDurationName,
		Help:    "gRPC unary request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "method"})
)

// Enabled — METRICS_ENABLED (default true).
func Enabled() bool {
	v := strings.TrimSpace(os.Getenv(obsenv.MetricsEnabled))
	if v == "" {
		return true
	}
	ok, err := strconv.ParseBool(v)
	return err == nil && ok
}

// RegisterHTTP регистрирует GET /metrics (Prometheus).
func RegisterHTTP(mux *http.ServeMux) {
	if !Enabled() {
		return
	}
	mux.Handle("GET "+obspath.Metrics, promhttp.Handler())
}

// RecordHTTP сохраняет метрики одного HTTP-запроса.
func RecordHTTP(service, method, path string, statusCode int, duration time.Duration) {
	if !Enabled() || obspath.IsProbe(path) {
		return
	}
	status := strconv.Itoa(statusCode)
	httpRequestsTotal.WithLabelValues(service, method, path, status).Inc()
	httpRequestDuration.WithLabelValues(service, method, path).Observe(duration.Seconds())
}

// RecordGRPC сохраняет метрики одного unary gRPC-вызова.
func RecordGRPC(service, method, code string, duration time.Duration) {
	if !Enabled() {
		return
	}
	grpcRequestsTotal.WithLabelValues(service, method, code).Inc()
	grpcRequestDuration.WithLabelValues(service, method).Observe(duration.Seconds())
}

