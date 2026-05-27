package grpcmw

import (
	"context"
	"log/slog"
	"testing"

	"github.com/dealer/dealer/pkg/obsenv"
	"github.com/dealer/dealer/pkg/obstest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const testRequest = "req"

func TestUnaryServerInterceptor(t *testing.T) {
	t.Setenv(obsenv.MetricsEnabled, obsenv.MetricsFalse)

	tests := []struct {
		name    string
		logger  *slog.Logger
		handler grpc.UnaryHandler
		wantErr error
		want    any
	}{
		{
			name:   "ok",
			logger: slog.Default(),
			handler: func(ctx context.Context, req any) (any, error) {
				if req != testRequest {
					t.Fatalf("req=%v", req)
				}
				return obstest.HealthBody, nil
			},
			want: obstest.HealthBody,
		},
		{
			name:   "grpc error",
			logger: slog.Default(),
			handler: func(context.Context, any) (any, error) {
				return nil, status.Error(codes.NotFound, "missing")
			},
			wantErr: status.Error(codes.NotFound, "missing"),
		},
		{
			name:   "nil logger",
			logger: nil,
			handler: func(context.Context, any) (any, error) {
				return nil, nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ic := UnaryServerInterceptor(obstest.ServiceName, tc.logger)
			resp, err := ic(context.Background(), testRequest, &grpc.UnaryServerInfo{
				FullMethod: obstest.GRPCFullMethod,
			}, tc.handler)
			if tc.wantErr != nil {
				if status.Code(err) != status.Code(tc.wantErr) {
					t.Fatalf("code=%v want %v", status.Code(err), status.Code(tc.wantErr))
				}
				return
			}
			if err != nil || resp != tc.want {
				t.Fatalf("resp=%v err=%v", resp, err)
			}
		})
	}
}
