//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

func newTestPart() *domain.Part {
	now := time.Now().UTC()
	return &domain.Part{
		ID:        uuid.New(),
		SKU:       "SKU-" + uuid.NewString(),
		Name:      "Деталь-" + uuid.NewString()[:8],
		Category:  "тормозная система",
		Quantity:  0,
		Unit:      "шт",
		Price:     "1500.00",
		Location:  "A-1-1",
		Notes:     "test part",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestPartRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewPartRepository(testPool)

	p := newTestPart()
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.SKU != p.SKU || got.Name != p.Name || got.Price != "1500.00" {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
	if got.FolderID != nil || got.BrandID != nil || got.WarehouseID != nil {
		t.Fatalf("nil refs expected, got %+v", got)
	}

	p.Name = "updated name"
	p.Price = "1800.50"
	p.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, p); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err = repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Name != "updated name" || got.Price != "1800.50" {
		t.Fatalf("update not persisted: %+v", got)
	}

	if err := repo.Delete(ctx, p.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, p.ID); err != pgx.ErrNoRows {
		t.Fatalf("expected ErrNoRows after delete, got %v", err)
	}
}

func TestPartRepository_UniqueSKU(t *testing.T) {
	ctx := context.Background()
	repo := NewPartRepository(testPool)

	p := newTestPart()
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	dup := newTestPart()
	dup.SKU = p.SKU
	err := repo.Create(ctx, dup)
	if err == nil {
		t.Fatal("expected unique violation on duplicate sku")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("expected SQLSTATE 23505, got %v", err)
	}
}

func TestPartRepository_ListFilters(t *testing.T) {
	ctx := context.Background()
	repo := NewPartRepository(testPool)

	folderID := uuid.New()
	brandID := uuid.New()
	warehouse := uuid.New()
	cat := "фильтры"
	if _, err := testPool.Exec(ctx,
		"INSERT INTO part_folders (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)",
		folderID, "Ф", time.Now().UTC(), time.Now().UTC()); err != nil {
		t.Fatalf("seed folder: %v", err)
	}
	p1 := newTestPart()
	p1.Category = cat
	p1.FolderID = &folderID
	p1.BrandID = &brandID
	p2 := newTestPart()
	p2.Category = cat
	p3 := newTestPart()
	for _, p := range []*domain.Part{p1, p2, p3} {
		if err := repo.Create(ctx, p); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	list, total, err := repo.List(ctx, domain.PartListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("List all: %v", err)
	}
	if int(total) < 3 || len(list) < 3 {
		t.Fatalf("List all: total=%d len=%d", total, len(list))
	}

	list, total, err = repo.List(ctx, domain.PartListFilter{Limit: 10, CategoryFilter: cat})
	if err != nil {
		t.Fatalf("List by category: %v", err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("List by category: total=%d len=%d", total, len(list))
	}

	list, total, err = repo.List(ctx, domain.PartListFilter{Limit: 10, Search: p1.Name})
	if err != nil {
		t.Fatalf("List by search: %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].ID != p1.ID {
		t.Fatalf("List by search: total=%d len=%d", total, len(list))
	}

	list, total, err = repo.List(ctx, domain.PartListFilter{Limit: 10, FolderID: &folderID})
	if err != nil {
		t.Fatalf("List by folder: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("List by folder: total=%d len=%d", total, len(list))
	}

	list, total, err = repo.List(ctx, domain.PartListFilter{Limit: 10, WarehouseID: &warehouse})
	if err != nil {
		t.Fatalf("List by warehouse (no stock): %v", err)
	}
	if total != 0 {
		t.Fatalf("List by warehouse expected 0, got %d", total)
	}

	if _, err := testPool.Exec(ctx,
		"INSERT INTO part_stock (part_id, warehouse_id, quantity, updated_at) VALUES ($1,$2,$3,now())",
		p1.ID, warehouse, 5); err != nil {
		t.Fatalf("seed stock: %v", err)
	}
	list, total, err = repo.List(ctx, domain.PartListFilter{Limit: 10, WarehouseID: &warehouse})
	if err != nil {
		t.Fatalf("List by warehouse (with stock): %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].ID != p1.ID {
		t.Fatalf("List by warehouse: total=%d len=%d", total, len(list))
	}

	list, _, err = repo.List(ctx, domain.PartListFilter{Limit: 1, Offset: 0})
	if err != nil {
		t.Fatalf("List paginated: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("pagination limit not honored, len=%d", len(list))
	}
}

func TestPartRepository_QuantityNegative(t *testing.T) {
	ctx := context.Background()
	repo := NewPartRepository(testPool)

	p := newTestPart()
	p.Quantity = -1
	if err := repo.Create(ctx, p); err == nil {
		t.Fatal("expected CHECK violation for negative quantity, got nil")
	}
}
