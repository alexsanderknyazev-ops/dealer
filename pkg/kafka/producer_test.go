package kafka

import (
	"context"
	"testing"
	"time"
)

func TestNewProducer(t *testing.T) {
	t.Parallel()
	p := NewProducer([]string{"127.0.0.1:9092"}, "test-topic")
	if p == nil {
		t.Fatal("expected non-nil producer")
	}
	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestProducer_Publish_ContextDeadline(t *testing.T) {
	t.Parallel()
	// Unreachable broker: Publish should return an error within timeout.
	p := NewProducer([]string{"127.0.0.1:1"}, "test-topic")
	defer func() { _ = p.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := p.Publish(ctx, []byte("key"), []byte("value"))
	if err == nil {
		t.Fatal("expected error publishing to unreachable broker")
	}
}
