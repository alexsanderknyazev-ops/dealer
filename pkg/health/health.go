package health

import (
	"context"
	"net/http"

	"github.com/dealer/dealer/pkg/obspath"
)

const bodyOK = "ok"

// Check — проверка готовности (например, ping БД).
type Check func(ctx context.Context) error

// Register добавляет GET /healthz (всегда 200) и GET /readyz (Check).
func Register(mux *http.ServeMux, ready Check) {
	mux.HandleFunc("GET "+obspath.Healthz, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(bodyOK))
	})
	mux.HandleFunc("GET "+obspath.Readyz, func(w http.ResponseWriter, r *http.Request) {
		if ready == nil {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(bodyOK))
			return
		}
		if err := ready(r.Context()); err != nil {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(bodyOK))
	})
}
