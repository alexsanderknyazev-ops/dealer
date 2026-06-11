package grpclient

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

type authKey struct{}

// WithAuthorization stores the raw Authorization header value for downstream gRPC calls.
func WithAuthorization(ctx context.Context, authorization string) context.Context {
	if authorization == "" {
		return ctx
	}
	return context.WithValue(ctx, authKey{}, authorization)
}

// OutgoingContext forwards authorization from incoming gRPC metadata or context to outbound calls.
func OutgoingContext(ctx context.Context) context.Context {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("authorization"); len(vals) > 0 {
			return metadata.AppendToOutgoingContext(ctx, "authorization", vals[0])
		}
	}
	if auth, ok := ctx.Value(authKey{}).(string); ok && auth != "" {
		val := auth
		if !strings.HasPrefix(strings.ToLower(val), "bearer ") {
			val = "Bearer " + strings.TrimPrefix(val, "Bearer ")
		}
		return metadata.AppendToOutgoingContext(ctx, "authorization", val)
	}
	return ctx
}
