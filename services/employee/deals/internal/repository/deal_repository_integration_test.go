//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/deals/internal/domain"
)

func newTestDeal() *domain.Deal {
	now := time.Now().UTC()
	return &domain.Deal{
		ID:         uuid.New(),
		CustomerID: uuid.New(),
		VehicleID:  uuid.New(),
		Amount:     "2500000.00",
		Stage:      "draft",
		Notes:      "test deal",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func TestDealRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewDealRepository(testPool)

	d := newTestDeal()
	if err := repo.Create(ctx, d); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, d.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Amount != d.Amount || got.Stage != d.Stage || got.Notes != d.Notes {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
	if got.AssignedTo != nil {
		t.Fatalf("assigned_to should be nil, got %v", got.AssignedTo)
	}

	assigned := uuid.New()
	d.Stage = "paid"
	d.AssignedTo = &assigned
	d.Notes = "updated"
	d.Amount = "3000000.00"
	d.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, d); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err = repo.GetByID(ctx, d.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Stage != "paid" || got.Amount != "3000000.00" || got.Notes != "updated" {
		t.Fatalf("update not persisted: %+v", got)
	}
	if got.AssignedTo == nil || *got.AssignedTo != assigned {
		t.Fatalf("assigned_to not persisted: %+v", got.AssignedTo)
	}

	if err := repo.Delete(ctx, d.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, d.ID); err != pgx.ErrNoRows {
		t.Fatalf("expected ErrNoRows after delete, got %v", err)
	}
}

func TestDealRepository_ListFilters(t *testing.T) {
	ctx := context.Background()
	repo := NewDealRepository(testPool)

	customer := uuid.New()
	d1 := newTestDeal()
	d1.CustomerID = customer
	d1.Stage = "completed"
	d2 := newTestDeal()
	d2.CustomerID = customer
	d2.Stage = "in_progress"
	d3 := newTestDeal()
	d3.Stage = "cancelled"
	for _, d := range []*domain.Deal{d1, d2, d3} {
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	list, total, err := repo.List(ctx, 10, 0, "", "")
	if err != nil {
		t.Fatalf("List all: %v", err)
	}
	if int(total) < 3 || len(list) < 3 {
		t.Fatalf("List all: total=%d len=%d", total, len(list))
	}

	list, total, err = repo.List(ctx, 10, 0, "completed", "")
	if err != nil {
		t.Fatalf("List by stage: %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].Stage != "completed" {
		t.Fatalf("List by stage: total=%d len=%d %+v", total, len(list), list[0])
	}

	list, total, err = repo.List(ctx, 10, 0, "", customer.String())
	if err != nil {
		t.Fatalf("List by customer: %v", err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("List by customer: total=%d len=%d", total, len(list))
	}

	list, _, err = repo.List(ctx, 1, 0, "", "")
	if err != nil {
		t.Fatalf("List paginated: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("pagination limit not honored, len=%d", len(list))
	}
}

func TestDealRepository_InvalidStage(t *testing.T) {
	ctx := context.Background()
	repo := NewDealRepository(testPool)

	d := newTestDeal()
	d.Stage = "not_a_stage"
	if err := repo.Create(ctx, d); err == nil {
		t.Fatal("expected CHECK violation for invalid stage, got nil")
	}
}
