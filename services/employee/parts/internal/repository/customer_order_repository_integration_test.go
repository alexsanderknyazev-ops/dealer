//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

func newTestCustomerOrder(partID, warehouseID uuid.UUID) *domain.CustomerOrder {
	now := time.Now().UTC()
	return &domain.CustomerOrder{
		ID:               uuid.New(),
		Status:           "draft",
		CustomerID:       uuid.New(),
		IssueWarehouseID: warehouseID,
		Notes:            "test customer order",
		CreatedAt:        now,
		UpdatedAt:        now,
		Lines: []domain.PartOrderLine{
			{
				ID:        uuid.New(),
				PartID:    partID,
				Quantity:  2,
				UnitPrice: "1500.00",
				SortOrder: 1,
				CreatedAt: now,
			},
		},
	}
}

func TestCustomerOrderRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	partRepo := NewPartRepository(testPool)
	repo := NewCustomerOrderRepository(testPool)

	p := newTestPart()
	if err := partRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create part: %v", err)
	}
	wh := uuid.New()

	o := newTestCustomerOrder(p.ID, wh)
	num, err := repo.NextOrderNumber(ctx)
	if err != nil {
		t.Fatalf("NextOrderNumber: %v", err)
	}
	o.OrderNumber = num
	if err := repo.Create(ctx, o); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, o.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.OrderNumber != num || got.CustomerID != o.CustomerID || len(got.Lines) != 1 {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
	if got.Lines[0].Quantity != 2 || got.Lines[0].UnitPrice != "1500.00" {
		t.Fatalf("line mismatch: %+v", got.Lines[0])
	}

	vehicleID := uuid.New()
	o.VehicleID = &vehicleID
	o.Notes = "updated"
	o.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, o, false); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ = repo.GetByID(ctx, o.ID)
	if got.VehicleID == nil || *got.VehicleID != vehicleID || got.Notes != "updated" {
		t.Fatalf("update not persisted: %+v", got)
	}

	o.Lines = []domain.PartOrderLine{
		{
			ID:        uuid.New(),
			PartID:    p.ID,
			Quantity:  7,
			UnitPrice: "1300.00",
			SortOrder: 1,
			CreatedAt: time.Now().UTC(),
		},
	}
	if err := repo.Update(ctx, o, true); err != nil {
		t.Fatalf("Update replaceLines: %v", err)
	}
	got, _ = repo.GetByID(ctx, o.ID)
	if len(got.Lines) != 1 || got.Lines[0].Quantity != 7 {
		t.Fatalf("replaceLines not honored: %+v", got.Lines)
	}
}

func TestCustomerOrderRepository_Lifecycle(t *testing.T) {
	ctx := context.Background()
	partRepo := NewPartRepository(testPool)
	repo := NewCustomerOrderRepository(testPool)

	p := newTestPart()
	if err := partRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create part: %v", err)
	}

	o := newTestCustomerOrder(p.ID, uuid.New())
	num, _ := repo.NextOrderNumber(ctx)
	o.OrderNumber = num
	if err := repo.Create(ctx, o); err != nil {
		t.Fatalf("Create: %v", err)
	}

	woID := uuid.New()
	if err := repo.LinkWorkOrder(ctx, o.ID, woID, time.Now().UTC()); err != nil {
		t.Fatalf("LinkWorkOrder: %v", err)
	}
	got, _ := repo.GetByID(ctx, o.ID)
	if got.Status != "linked" || got.FulfillmentWorkOrderID == nil || *got.FulfillmentWorkOrderID != woID {
		t.Fatalf("LinkWorkOrder not persisted: %+v", got)
	}

	if err := repo.MarkFulfilled(ctx, o.ID); err != nil {
		t.Fatalf("MarkFulfilled: %v", err)
	}
	got, _ = repo.GetByID(ctx, o.ID)
	if got.Status != "fulfilled" {
		t.Fatalf("expected fulfilled, got %s", got.Status)
	}
}

func TestCustomerOrderRepository_List(t *testing.T) {
	ctx := context.Background()
	partRepo := NewPartRepository(testPool)
	repo := NewCustomerOrderRepository(testPool)

	p := newTestPart()
	if err := partRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create part: %v", err)
	}

	o := newTestCustomerOrder(p.ID, uuid.New())
	num, _ := repo.NextOrderNumber(ctx)
	o.OrderNumber = num
	if err := repo.Create(ctx, o); err != nil {
		t.Fatalf("Create: %v", err)
	}

	list, total, err := repo.List(ctx, 10, 0, "draft")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total < 1 || len(list) < 1 {
		t.Fatalf("List: total=%d len=%d", total, len(list))
	}
}
