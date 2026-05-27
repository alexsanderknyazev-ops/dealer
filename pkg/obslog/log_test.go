package obslog

import (
	"context"
	"log/slog"
	"testing"

	"github.com/dealer/dealer/pkg/obsenv"
	"github.com/dealer/dealer/pkg/obstest"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		wantInfo  bool
		wantError bool
	}{
		{"default info", "", true, true},
		{"error level", "error", false, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(obsenv.LogLevel, tc.logLevel)
			t.Setenv(obsenv.ServiceVersion, obstest.TestVersion)
			logger := Init(obstest.ServiceObslog)
			if logger == nil || Default != logger {
				t.Fatal("logger not initialized")
			}
			ctx := context.Background()
			if logger.Enabled(ctx, slog.LevelInfo) != tc.wantInfo {
				t.Fatalf("info enabled=%v want %v", logger.Enabled(ctx, slog.LevelInfo), tc.wantInfo)
			}
			if logger.Enabled(ctx, slog.LevelError) != tc.wantError {
				t.Fatalf("error enabled=%v want %v", logger.Enabled(ctx, slog.LevelError), tc.wantError)
			}
		})
	}
}
