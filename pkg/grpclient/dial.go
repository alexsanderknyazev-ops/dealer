package grpclient

import (
	"context"

	"google.golang.org/grpc"
)

// Dial opens an insecure gRPC client connection (cluster-internal use).
func Dial(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	opts := append(DefaultDialOptions(), grpc.WithBlock())
	return grpc.DialContext(ctx, addr, opts...)
}
