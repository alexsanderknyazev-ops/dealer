//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/dealerpoints/internal/domain"
)

func TestDealerPointRepository_CRUD(t *testing.T) {
	repo := NewDealerPointRepository(testPool)
	ctx := context.Background()

	d := &domain.DealerPoint{
		ID:        uuid.New(),
		Name:      "IT Dealer Point",
		Address:   "Moscow",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, d); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, d.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != d.Name || got.Address != d.Address {
		t.Fatalf("GetByID mismatch: %+v", got)
	}

	if _, err := repo.GetByID(ctx, uuid.New()); err != pgx.ErrNoRows {
		t.Fatalf("GetByID(missing): got %v want pgx.ErrNoRows", err)
	}

	d.Name = "IT Dealer Point Renamed"
	d.Address = "Saint Petersburg"
	d.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, d); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err = repo.GetByID(ctx, d.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Name != "IT Dealer Point Renamed" || got.Address != "Saint Petersburg" {
		t.Fatalf("Update not persisted: %+v", got)
	}

	if err := repo.Delete(ctx, d.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, d.ID); err != pgx.ErrNoRows {
		t.Fatalf("GetByID after delete: got %v want pgx.ErrNoRows", err)
	}
}

func TestDealerPointRepository_ListSearchAndPagination(t *testing.T) {
	repo := NewDealerPointRepository(testPool)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		d := &domain.DealerPoint{
			ID:        uuid.New(),
			Name:      fmt.Sprintf("IT DP Point %d", i),
			Address:   fmt.Sprintf("Lenina str %d", i),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	list, total, err := repo.List(ctx, 10, 0, "IT DP Point")
	if err != nil {
		t.Fatalf("List search: %v", err)
	}
	if total < 3 || len(list) < 3 {
		t.Fatalf("List search: total=%d len=%d want >=3", total, len(list))
	}

	list, total, err = repo.List(ctx, 10, 0, "Lenina str")
	if err != nil {
		t.Fatalf("List address search: %v", err)
	}
	if total < 3 {
		t.Fatalf("List address search total: got %d want >=3", total)
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
