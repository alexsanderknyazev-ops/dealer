//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestSupplierRepository_ListAndLookup(t *testing.T) {
	ctx := context.Background()
	repo := NewSupplierRepository(testPool)

	supplierID, _ := uuid.Parse("a8800001-0000-4000-8000-000000000001")
	exists, err := repo.Exists(ctx, supplierID)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !exists {
		t.Fatal("seeded supplier should exist")
	}
	if _, err := repo.Exists(ctx, uuid.New()); err != nil {
		t.Fatalf("Exists random: %v", err)
	}

	name := repo.Name(ctx, supplierID)
	if name == "" {
		t.Fatal("seeded supplier name should not be empty")
	}
	if repo.Name(ctx, uuid.New()) != "" {
		t.Fatal("unknown supplier name should be empty")
	}

	list, total, err := repo.List(ctx, 10, 0, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total < 2 || len(list) < 2 {
		t.Fatalf("expected at least 2 seeded suppliers, total=%d len=%d", total, len(list))
	}

	list, total, err = repo.List(ctx, 10, 0, "АвтоПоставка")
	if err != nil {
		t.Fatalf("List search: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("List search: total=%d len=%d", total, len(list))
	}

	list, _, err = repo.List(ctx, 0, 0, "")
	if err != nil {
		t.Fatalf("List default limit: %v", err)
	}
	if len(list) > 100 {
		t.Fatalf("default limit 100 violated, got %d", len(list))
	}
}
