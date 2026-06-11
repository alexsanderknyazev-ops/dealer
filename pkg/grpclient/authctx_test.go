package grpclient

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestOutgoingContext_FromIncomingMetadata(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer tok"))
	out := OutgoingContext(ctx)
	md, ok := metadata.FromOutgoingContext(out)
	if !ok || len(md.Get("authorization")) != 1 || md.Get("authorization")[0] != "Bearer tok" {
		t.Fatalf("metadata: %v ok=%v", md, ok)
	}
}

func TestOutgoingContext_FromContextValue(t *testing.T) {
	ctx := WithAuthorization(context.Background(), "Bearer abc")
	out := OutgoingContext(ctx)
	md, ok := metadata.FromOutgoingContext(out)
	if !ok || md.Get("authorization")[0] != "Bearer abc" {
		t.Fatalf("metadata: %v ok=%v", md, ok)
	}
}
