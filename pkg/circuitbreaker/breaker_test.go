package circuitbreaker

import (
	"errors"
	"testing"
	"time"
)

func TestBreakerOpensAfterFailures(t *testing.T) {
	b := New(Config{FailureThreshold: 3, SuccessThreshold: 1, OpenTimeout: time.Minute})
	for i := 0; i < 2; i++ {
		if err := b.Allow(); err != nil {
			t.Fatalf("allow %d: %v", i, err)
		}
		b.RecordFailure()
	}
	if b.State() != StateClosed {
		t.Fatalf("expected closed, got %s", b.State())
	}
	b.RecordFailure()
	if b.State() != StateOpen {
		t.Fatalf("expected open, got %s", b.State())
	}
	if !errors.Is(b.Allow(), ErrOpen) {
		t.Fatal("expected ErrOpen")
	}
}

func TestBreakerHalfOpenThenCloses(t *testing.T) {
	b := New(Config{FailureThreshold: 1, SuccessThreshold: 2, OpenTimeout: 10 * time.Millisecond})
	b.RecordFailure()
	if b.State() != StateOpen {
		t.Fatalf("expected open, got %s", b.State())
	}
	time.Sleep(15 * time.Millisecond)
	if b.State() != StateHalfOpen {
		t.Fatalf("expected half_open, got %s", b.State())
	}
	b.RecordSuccess()
	if b.State() != StateHalfOpen {
		t.Fatalf("expected half_open after one success, got %s", b.State())
	}
	b.RecordSuccess()
	if b.State() != StateClosed {
		t.Fatalf("expected closed, got %s", b.State())
	}
}

func TestRegistryReturnsSameBreaker(t *testing.T) {
	r := NewRegistry(DefaultConfig())
	a := r.Get("vehicles-service:50053")
	b := r.Get("vehicles-service:50053")
	if a != b {
		t.Fatal("expected same breaker instance")
	}
}
