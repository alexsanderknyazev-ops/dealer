package grpclient

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ExistsFromRPC treats gRPC NotFound as "does not exist", other errors propagate.
func ExistsFromRPC(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
		return false, nil
	}
	return false, err
}
