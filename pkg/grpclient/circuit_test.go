package grpclient

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type unavailableService struct {
	grpc.ServiceRegistrar
}

func (unavailableService) Failing(context.Context, *struct{}, *struct{}) error {
	return status.Error(codes.Unavailable, "down")
}

func TestUnaryClientInterceptorOpensOnUnavailable(t *testing.T) {
	t.Setenv("CIRCUIT_BREAKER_ENABLED", "true")
	t.Setenv("CIRCUIT_BREAKER_FAILURE_THRESHOLD", "2")

	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	grpc.RegisterService(&grpc.ServiceDesc{
		ServiceName: "test.Unavailable",
		HandlerType: (*unavailableService)(nil),
		Methods: []grpc.MethodDesc{{
			MethodName: "Failing",
			Handler: func(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
				return nil, status.Error(codes.Unavailable, "down")
			},
		}},
	}, unavailableService{})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop() })

	dialer := func(context.Context, string) (net.Conn, error) { return lis.Dial() }
	conn, err := grpc.NewClient("passthrough:///test",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(UnaryClientInterceptor()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	invoke := func(ctx context.Context) error {
		return conn.Invoke(ctx, "/test.Unavailable/Failing", &struct{}{}, &struct{}{})
	}

	for i := 0; i < 2; i++ {
		if err := invoke(context.Background()); status.Code(err) != codes.Unavailable {
			t.Fatalf("call %d: expected unavailable, got %v", i, err)
		}
	}
	err = invoke(context.Background())
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("expected unavailable when open, got %v", err)
	}
	if !contains(err.Error(), "circuit breaker open") {
		t.Fatalf("expected circuit breaker message, got %v", err)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
