//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/dealerpoints/internal/domain"
)

func TestWarehouseRepository_CRUD(t *testing.T) {
	dpRepo := NewDealerPointRepository(testPool)
	leRepo := NewLegalEntityRepository(testPool)
	ctx := context.Background()

	dp := &domain.DealerPoint{ID: uuid.New(), Name: "IT DP Warehouse", Address: "Moscow", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := dpRepo.Create(ctx, dp); err != nil {
		t.Fatalf("Create dealer point: %v", err)
	}
	le := &domain.LegalEntity{ID: uuid.New(), Name: "IT LE Warehouse", INN: "7700112233", Address: "Moscow", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := leRepo.Create(ctx, le); err != nil {
		t.Fatalf("Create legal entity: %v", err)
	}

	repo := NewWarehouseRepository(testPool)
	w := &domain.Warehouse{
		ID:            uuid.New(),
		DealerPointID: dp.ID,
		LegalEntityID: le.ID,
		Type:          "cars",
		Name:          "IT Main Warehouse",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	if err := repo.Create(ctx, w); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, w.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != w.Name || got.Type != w.Type || got.DealerPointID != dp.ID || got.LegalEntityID != le.ID {
		t.Fatalf("GetByID mismatch: %+v", got)
	}

	if _, err := repo.GetByID(ctx, uuid.New()); err != pgx.ErrNoRows {
		t.Fatalf("GetByID(missing): got %v want pgx.ErrNoRows", err)
	}

	w.Name = "IT Main Warehouse Renamed"
	w.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, w); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err = repo.GetByID(ctx, w.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Name != "IT Main Warehouse Renamed" {
		t.Fatalf("Update not persisted: %+v", got)
	}

	if err := repo.Delete(ctx, w.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, w.ID); err != pgx.ErrNoRows {
		t.Fatalf("GetByID after delete: got %v want pgx.ErrNoRows", err)
	}
}

func TestWarehouseRepository_ListFilters(t *testing.T) {
	dpRepo := NewDealerPointRepository(testPool)
	leRepo := NewLegalEntityRepository(testPool)
	ctx := context.Background()

	dp := &domain.DealerPoint{ID: uuid.New(), Name: "IT DP Warehouse List", Address: "Moscow", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := dpRepo.Create(ctx, dp); err != nil {
		t.Fatalf("Create dealer point: %v", err)
	}
	le := &domain.LegalEntity{ID: uuid.New(), Name: "IT LE Warehouse List", INN: "7700223344", Address: "Moscow", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := leRepo.Create(ctx, le); err != nil {
		t.Fatalf("Create legal entity: %v", err)
	}

	repo := NewWarehouseRepository(testPool)
	warehouses := []*domain.Warehouse{
		{ID: uuid.New(), DealerPointID: dp.ID, LegalEntityID: le.ID, Type: "cars", Name: "IT WH Cars", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		{ID: uuid.New(), DealerPointID: dp.ID, LegalEntityID: le.ID, Type: "parts", Name: "IT WH Parts", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
	}
	for _, w := range warehouses {
		if err := repo.Create(ctx, w); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	list, total, err := repo.List(ctx, 10, 0, &dp.ID, nil, "")
	if err != nil {
		t.Fatalf("List dealer point: %v", err)
	}
	if total < 2 {
		t.Fatalf("List dealer point total: got %d want >=2", total)
	}

	list, total, err = repo.List(ctx, 10, 0, nil, nil, "parts")
	if err != nil {
		t.Fatalf("List type: %v", err)
	}
	if total < 1 {
		t.Fatalf("List type total: got %d want >=1", total)
	}
	for _, w := range list {
		if w.Type != "parts" {
			t.Fatalf("List type filter returned %q", w.Type)
		}
	}

	list, total, err = repo.List(ctx, 10, 0, &dp.ID, &le.ID, "cars")
	if err != nil {
		t.Fatalf("List combined: %v", err)
	}
	if total < 1 {
		t.Fatalf("List combined total: got %d want >=1", total)
	}

	page, total, err := repo.List(ctx, 1, 0, nil, nil, "")
	if err != nil {
		t.Fatalf("List paged: %v", err)
	}
	if len(page) > 1 {
		t.Fatalf("List page len: got %d want <=1", len(page))
	}
	if total < 2 {
		t.Fatalf("List page total: got %d want >=2", total)
	}
}
