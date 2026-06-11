package httplog

import (
	"net/http"
	"testing"
)

func TestShouldReportHTTP(t *testing.T) {
	tests := []struct {
		status int
		want   bool
	}{
		{http.StatusOK, false},
		{http.StatusNotFound, false},
		{http.StatusUnauthorized, false},
		{http.StatusBadRequest, false},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
	}
	for _, tc := range tests {
		if got := shouldReportHTTP(tc.status); got != tc.want {
			t.Fatalf("status %d: got %v want %v", tc.status, got, tc.want)
		}
	}
}
