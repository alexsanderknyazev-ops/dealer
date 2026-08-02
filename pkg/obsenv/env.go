package obsenv

// Имена переменных окружения observability.
const (
	MetricsEnabled = "METRICS_ENABLED"
	LogLevel       = "LOG_LEVEL"
	ServiceVersion = "SERVICE_VERSION"
	OtlpEndpoint   = "OTEL_EXPORTER_OTLP_ENDPOINT"
	TracesExporter = "OTEL_TRACES_EXPORTER"
)

// Значения METRICS_ENABLED.
const (
	MetricsTrue  = "true"
	MetricsFalse = "false"
)
