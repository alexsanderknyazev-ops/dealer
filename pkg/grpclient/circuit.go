package grpclient

import (
	"context"
	"errors"

	"github.com/dealer/dealer/pkg/circuitbreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryClientInterceptor защищает исходящие gRPC-вызовы circuit breaker'ом по target.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	reg := circuitbreaker.DefaultRegistry()
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if !circuitbreaker.Enabled() {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		target := cc.Target()
		br := reg.Get(target)
		if err := br.Allow(); err != nil {
			return status.Errorf(codes.Unavailable, "circuit breaker open for %s", target)
		}
		err := invoker(ctx, method, req, reply, cc, opts...)
		if isBreakerFailure(err) {
			br.RecordFailure()
		} else {
			br.RecordSuccess()
		}
		return err
	}
}

func isBreakerFailure(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}
	switch status.Code(err) {
	case codes.Unavailable, codes.DeadlineExceeded, codes.Unknown, codes.ResourceExhausted, codes.Aborted:
		return true
	default:
		return false
	}
}
