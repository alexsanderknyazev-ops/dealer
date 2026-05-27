package obslog

import (
	"log/slog"
	"os"
	"strings"

	"github.com/dealer/dealer/pkg/obsenv"
)

// Default — логгер сервиса после Init.
var Default = slog.Default()

// Init настраивает JSON-лог в stdout с полями service и version (SERVICE_VERSION).
func Init(service string) *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(strings.TrimSpace(os.Getenv(obsenv.LogLevel))) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})
	attrs := []any{slog.String("service", service)}
	if v := strings.TrimSpace(os.Getenv(obsenv.ServiceVersion)); v != "" {
		attrs = append(attrs, slog.String("version", v))
	}
	logger := slog.New(handler).With(attrs...)
	slog.SetDefault(logger)
	Default = logger
	return logger
}
