package httplog

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dealer/dealer/pkg/obsenv"
	"github.com/dealer/dealer/pkg/obspath"
	"github.com/dealer/dealer/pkg/obstest"
)

type captureHandler struct{ records []slog.Record }

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}
func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler      { return h }

var _ slog.Handler = (*captureHandler)(nil)

func wrapWithStatus(t *testing.T, status int, logger *slog.Logger) (*captureHandler, *httptest.ResponseRecorder) {
	t.Helper()
	t.Setenv(obsenv.MetricsEnabled, obsenv.MetricsFalse)
	capture := &captureHandler{}
	if logger == nil {
		logger = slog.New(capture)
	}
	h := Wrap(obstest.ServiceName, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	}), logger, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, obstest.APITestPath, nil))
	return capture, rec
}

func TestWrap(t *testing.T) {
	t.Run("logs info", func(t *testing.T) {
		capture, rec := wrapWithStatus(t, http.StatusCreated, nil)
		if rec.Code != http.StatusCreated || len(capture.records) != 1 || capture.records[0].Level != slog.LevelInfo {
			t.Fatalf("code=%d records=%d", rec.Code, len(capture.records))
		}
	})

	t.Run("logs warn on 5xx", func(t *testing.T) {
		capture, _ := wrapWithStatus(t, http.StatusInternalServerError, nil)
		if len(capture.records) != 2 || capture.records[1].Level != slog.LevelWarn {
			t.Fatalf("records=%d", len(capture.records))
		}
	})

	t.Run("skips probe paths", func(t *testing.T) {
		t.Setenv(obsenv.MetricsEnabled, obsenv.MetricsFalse)
		capture := &captureHandler{}
		h := Wrap(obstest.ServiceName, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), slog.New(capture), nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, obspath.Healthz, nil))
		if rec.Code != http.StatusOK || len(capture.records) != 0 {
			t.Fatalf("code=%d records=%d", rec.Code, len(capture.records))
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		t.Setenv(obsenv.MetricsEnabled, obsenv.MetricsFalse)
		rec := httptest.NewRecorder()
		Wrap(obstest.ServiceName, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), nil, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, obstest.OKPath, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("json contains path", func(t *testing.T) {
		t.Setenv(obsenv.MetricsEnabled, obsenv.MetricsFalse)
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
		Wrap(obstest.ServiceName, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), logger, nil).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, obstest.APIItemsPath, nil))
		if !strings.Contains(buf.String(), obstest.APIItemsPath) {
			t.Fatalf("log: %s", buf.String())
		}
	})
}
