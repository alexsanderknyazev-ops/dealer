//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

func newTestSupplierOrder(partID, supplierID, warehouseID uuid.UUID) *domain.SupplierOrder {
	now := time.Now().UTC()
	return &domain.SupplierOrder{
		ID:                 uuid.New(),
		Status:             "draft",
		SupplierID:         supplierID,
		ReceiptWarehouseID: warehouseID,
		Notes:              "test supplier order",
		CreatedAt:          now,
		UpdatedAt:          now,
		Lines: []domain.PartOrderLine{
			{
				ID:        uuid.New(),
				PartID:    partID,
				Quantity:  10,
				UnitPrice: "1100.00",
				SortOrder: 1,
				CreatedAt: now,
			},
		},
	}
}

func TestSupplierOrderRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	partRepo := NewPartRepository(testPool)
	repo := NewSupplierOrderRepository(testPool)

	p := newTestPart()
	if err := partRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create part: %v", err)
	}
	supplierID, _ := uuid.Parse("a8800001-0000-4000-8000-000000000001")
	wh := uuid.New()

	o := newTestSupplierOrder(p.ID, supplierID, wh)
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
	if got.OrderNumber != num || got.SupplierID != supplierID || len(got.Lines) != 1 {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
	if got.Lines[0].UnitPrice != "1100.00" || got.Lines[0].Quantity != 10 {
		t.Fatalf("line mismatch: %+v", got.Lines[0])
	}

	newSupplierID, _ := uuid.Parse("a8800001-0000-4000-8000-000000000002")
	o.SupplierID = newSupplierID
	o.Notes = "updated"
	o.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, o, false); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ = repo.GetByID(ctx, o.ID)
	if got.SupplierID != newSupplierID || got.Notes != "updated" {
		t.Fatalf("update not persisted: %+v", got)
	}

	o.Lines = []domain.PartOrderLine{
		{
			ID:        uuid.New(),
			PartID:    p.ID,
			Quantity:  3,
			UnitPrice: "900.00",
			SortOrder: 1,
			CreatedAt: time.Now().UTC(),
		},
	}
	if err := repo.Update(ctx, o, true); err != nil {
		t.Fatalf("Update replaceLines: %v", err)
	}
	got, _ = repo.GetByID(ctx, o.ID)
	if len(got.Lines) != 1 || got.Lines[0].Quantity != 3 {
		t.Fatalf("replaceLines not honored: %+v", got.Lines)
	}
}

func TestSupplierOrderRepository_Lifecycle(t *testing.T) {
	ctx := context.Background()
	partRepo := NewPartRepository(testPool)
	repo := NewSupplierOrderRepository(testPool)

	p := newTestPart()
	if err := partRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create part: %v", err)
	}
	supplierID, _ := uuid.Parse("a8800001-0000-4000-8000-000000000001")

	o := newTestSupplierOrder(p.ID, supplierID, uuid.New())
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

	if err := repo.UpdateStatus(ctx, o.ID, "cancelled", nil, time.Now().UTC()); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	got, _ = repo.GetByID(ctx, o.ID)
	if got.Status != "cancelled" {
		t.Fatalf("expected cancelled, got %s", got.Status)
	}
}

func TestSupplierOrderRepository_List(t *testing.T) {
	ctx := context.Background()
	partRepo := NewPartRepository(testPool)
	repo := NewSupplierOrderRepository(testPool)

	p := newTestPart()
	if err := partRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create part: %v", err)
	}
	supplierID, _ := uuid.Parse("a8800001-0000-4000-8000-000000000001")

	o := newTestSupplierOrder(p.ID, supplierID, uuid.New())
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
	list, _, err = repo.List(ctx, 10, 0, "cancelled")
	if err != nil {
		t.Fatalf("List cancelled: %v", err)
	}
	for _, it := range list {
		if it.ID == o.ID {
			t.Fatal("draft order should not appear in cancelled list")
		}
	}
}
