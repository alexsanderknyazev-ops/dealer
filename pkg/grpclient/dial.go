package grpclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// Dial opens an insecure gRPC client connection (cluster-internal use)
// and blocks until the connection becomes Ready (or ctx is done).
func Dial(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, DefaultDialOptions()...)
	if err != nil {
		return nil, err
	}
	conn.Connect()
	for conn.GetState() != connectivity.Ready {
		if !conn.WaitForStateChange(ctx, conn.GetState()) {
			_ = conn.Close()
			return nil, fmt.Errorf("grpc dial %s: %w", addr, ctx.Err())
		}
	}
	return conn, nil
}
