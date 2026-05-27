package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dealer/dealer/pkg/obspath"
	"github.com/dealer/dealer/pkg/obstest"
)

func serveGET(t *testing.T, path string, ready Check) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	Register(mux, ready)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestHealthz(t *testing.T) {
	rec := serveGET(t, obspath.Healthz, nil)
	if rec.Code != http.StatusOK || rec.Body.String() != obstest.HealthBody {
		t.Fatalf("status=%d body=%q", rec.Code, rec.Body.String())
	}
}

func TestReadyz(t *testing.T) {
	tests := []struct {
		name  string
		ready Check
		want  int
	}{
		{"nil check", nil, http.StatusOK},
		{"ok", func(context.Context) error { return nil }, http.StatusOK},
		{"fail", func(context.Context) error { return errors.New("db down") }, http.StatusServiceUnavailable},
		{"postgres ok", Postgres(&fakePinger{}), http.StatusOK},
		{"postgres fail", Postgres(&fakePinger{err: errors.New("ping failed")}), http.StatusServiceUnavailable},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := serveGET(t, obspath.Readyz, tc.ready)
			if rec.Code != tc.want {
				t.Fatalf("status=%d want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestPostgres_NilPool(t *testing.T) {
	if Postgres(nil) != nil {
		t.Fatal("expected nil check")
	}
}

type fakePinger struct{ err error }

func (f *fakePinger) Ping(context.Context) error { return f.err }
