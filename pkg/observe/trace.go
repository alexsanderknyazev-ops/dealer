package observe

import (
	"context"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"github.com/dealer/dealer/pkg/obsenv"
)

// InitTracing настраивает OpenTelemetry: экспорт трейсов в OTLP/HTTP.
// Коллектор берётся из OTEL_EXPORTER_OTLP_ENDPOINT; если переменная не задана
// (или OTEL_TRACES_EXPORTER=none) — трейсинг no-op. Возвращает shutdown для flush.
func InitTracing(service string) func() {
	endpoint := os.Getenv(obsenv.OtlpEndpoint)
	if endpoint == "" || strings.EqualFold(os.Getenv(obsenv.TracesExporter), "none") {
		return func() {}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint))
	if err != nil {
		return func() {}
	}
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(service),
			semconv.ServiceVersion(os.Getenv(obsenv.ServiceVersion)),
		),
	)
	if err != nil {
		return func() {}
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = provider.Shutdown(shutdownCtx)
	}
}
