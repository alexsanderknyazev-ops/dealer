package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dealer/dealer/pkg/obsenv"
	"github.com/dealer/dealer/pkg/obspath"
	"github.com/dealer/dealer/pkg/obstest"
)

func TestEnabled(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{"", true},
		{obsenv.MetricsFalse, false},
		{obsenv.MetricsTrue, true},
	}
	for _, tc := range tests {
		t.Run(tc.env, func(t *testing.T) {
			t.Setenv(obsenv.MetricsEnabled, tc.env)
			if Enabled() != tc.want {
				t.Fatalf("got %v want %v", Enabled(), tc.want)
			}
		})
	}
}

func TestRegisterHTTP(t *testing.T) {
	tests := []struct {
		name       string
		enabled    string
		record     bool
		wantStatus int
		wantMetric bool
	}{
		{"enabled", obsenv.MetricsTrue, true, http.StatusOK, true},
		{"disabled", obsenv.MetricsFalse, false, http.StatusNotFound, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(obsenv.MetricsEnabled, tc.enabled)
			if tc.record {
				RecordHTTP(obstest.ServiceName, http.MethodGet, obstest.APIItemsPath, http.StatusOK, time.Millisecond)
			}
			mux := http.NewServeMux()
			RegisterHTTP(mux)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, obspath.Metrics, nil))
			if rec.Code != tc.wantStatus {
				t.Fatalf("status=%d", rec.Code)
			}
			if tc.wantMetric && !strings.Contains(rec.Body.String(), HTTPRequestsTotalName) {
				t.Fatal("missing " + HTTPRequestsTotalName)
			}
		})
	}
}

func TestRecordHTTP_SkipsProbePaths(t *testing.T) {
	t.Setenv(obsenv.MetricsEnabled, obsenv.MetricsTrue)
	for _, path := range obspath.Probes {
		RecordHTTP(obstest.ServiceName, http.MethodGet, path, http.StatusOK, time.Millisecond)
	}
}

func TestRecordGRPC_WhenDisabled(t *testing.T) {
	t.Setenv(obsenv.MetricsEnabled, obsenv.MetricsFalse)
	RecordGRPC(obstest.ServiceName, obstest.GRPCMethodShort, obstest.GRPCCodeOK, time.Millisecond)
}
