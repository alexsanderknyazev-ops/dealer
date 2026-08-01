//go:build integration

package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/dealer/dealer/services/brands/internal/domain"
)

func TestBrandRepository_CRUD(t *testing.T) {
	repo := NewBrandRepository(testPool)
	ctx := context.Background()

	b := &domain.Brand{
		ID:        uuid.New(),
		Name:      "IT Brand",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, b); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, b.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != b.Name {
		t.Fatalf("GetByID mismatch: %+v", got)
	}

	if _, err := repo.GetByID(ctx, uuid.New()); err != pgx.ErrNoRows {
		t.Fatalf("GetByID(missing): got %v want pgx.ErrNoRows", err)
	}

	b.Name = "IT Brand Renamed"
	b.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, b); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err = repo.GetByID(ctx, b.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Name != "IT Brand Renamed" {
		t.Fatalf("Update not persisted: %+v", got)
	}

	if err := repo.Delete(ctx, b.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, b.ID); err != pgx.ErrNoRows {
		t.Fatalf("GetByID after delete: got %v want pgx.ErrNoRows", err)
	}
}

func TestBrandRepository_ListSearchAndPagination(t *testing.T) {
	repo := NewBrandRepository(testPool)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		b := &domain.Brand{
			ID:        uuid.New(),
			Name:      fmt.Sprintf("IT Search Brand %d", i),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		if err := repo.Create(ctx, b); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	list, total, err := repo.List(ctx, 10, 0, "IT Search Brand")
	if err != nil {
		t.Fatalf("List search: %v", err)
	}
	if total < 3 || len(list) < 3 {
		t.Fatalf("List search: total=%d len=%d want >=3", total, len(list))
	}

	page, total, err := repo.List(ctx, 2, 0, "")
	if err != nil {
		t.Fatalf("List paged: %v", err)
	}
	if len(page) > 2 {
		t.Fatalf("List page len: got %d want <=2", len(page))
	}
	if total < 3 {
		t.Fatalf("List page total: got %d want >=3", total)
	}
}

func TestBrandRepository_UniqueName(t *testing.T) {
	repo := NewBrandRepository(testPool)
	ctx := context.Background()

	b := &domain.Brand{
		ID:        uuid.New(),
		Name:      "IT Unique Brand",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, b); err != nil {
		t.Fatalf("Create: %v", err)
	}

	dup := *b
	dup.ID = uuid.New()
	dup.Name = "  it UNIQUE brand  "
	err := repo.Create(ctx, &dup)
	if err == nil {
		t.Fatal("expected unique violation for duplicate name")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("expected SQLSTATE 23505, got %v", err)
	}
}
