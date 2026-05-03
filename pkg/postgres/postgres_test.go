package postgres

import (
	"context"
	"testing"
)

func TestNewPool_InvalidDSN(t *testing.T) {
	t.Parallel()
	_, err := NewPool(context.Background(), "not-a-valid-postgres-dsn")
	if err == nil {
		t.Fatal("expected error for invalid DSN")
	}
}

func TestNewPool_InvalidURLScheme(t *testing.T) {
	t.Parallel()
	_, err := NewPool(context.Background(), "mysql://user:pass@localhost:3306/db")
	if err == nil {
		t.Fatal("expected error for non-postgres DSN")
	}
}
