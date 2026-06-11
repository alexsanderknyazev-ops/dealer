package observe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dealer/dealer/pkg/obsenv"
	"github.com/dealer/dealer/pkg/obspath"
	"github.com/dealer/dealer/pkg/obstest"
	"google.golang.org/grpc"
)

func TestObserve_HTTPAndGRPC(t *testing.T) {
	t.Setenv(obsenv.MetricsEnabled, obsenv.MetricsTrue)
	logger := Init(obstest.ServiceObserve)
	if logger == nil {
		t.Fatal("nil logger")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET "+obstest.APIPingPath, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	RegisterHTTP(mux, func(context.Context) error { return nil })

	rec := httptest.NewRecorder()
	WrapHTTP(obstest.ServiceObserve, mux, logger).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, obstest.APIPingPath, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("api status=%d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, obspath.Metrics, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics status=%d", rec.Code)
	}

	opts := GRPCServerOptions(obstest.ServiceObserve, logger, nil)
	if len(opts) != 1 {
		t.Fatalf("opts=%d", len(opts))
	}
	if grpc.NewServer(opts...) == nil {
		t.Fatal("nil server")
	}
}
