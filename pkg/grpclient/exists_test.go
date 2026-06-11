package grpclient

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestExistsFromRPC(t *testing.T) {
	ok, err := ExistsFromRPC(nil)
	if err != nil || !ok {
		t.Fatalf("nil: ok=%v err=%v", ok, err)
	}
	ok, err = ExistsFromRPC(status.Error(codes.NotFound, "missing"))
	if err != nil || ok {
		t.Fatalf("not found: ok=%v err=%v", ok, err)
	}
	ok, err = ExistsFromRPC(status.Error(codes.Internal, "boom"))
	if err == nil || ok {
		t.Fatalf("internal: ok=%v err=%v", ok, err)
	}
}
