package httplog

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/dealer/dealer/pkg/metrics"
	"github.com/dealer/dealer/pkg/obspath"
)

// Wrap — логирование запросов + метрики (если включены).
func Wrap(service string, next http.Handler, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if obspath.IsProbe(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		rw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		metrics.RecordHTTP(service, r.Method, r.URL.Path, rw.status, time.Since(start))
		logger.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote", r.RemoteAddr,
		)
		if rw.status >= http.StatusInternalServerError {
			logger.Warn("http server error response",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
			)
		}
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
