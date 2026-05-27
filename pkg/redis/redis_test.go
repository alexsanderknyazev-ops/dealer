package redis

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func TestNewClient(t *testing.T) {
	t.Parallel()
	c := NewClient("127.0.0.1:6379", "", 0)
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	_ = c.Close()
}

func TestPing_OK(t *testing.T) {
	t.Parallel()
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	c := NewClient(s.Addr(), "", 0)
	defer c.Close()

	if err := Ping(context.Background(), c); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestPing_Err(t *testing.T) {
	t.Parallel()
	// Nothing listening on this port on localhost.
	c := NewClient("127.0.0.1:1", "", 0)
	defer c.Close()

	ctx := context.Background()
	if err := Ping(ctx, c); err == nil {
		t.Fatal("expected error when Redis is unreachable")
	}
}
