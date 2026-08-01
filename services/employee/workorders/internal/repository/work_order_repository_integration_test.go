//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/workorders/internal/domain"
)

func newTestWorkOrder() *domain.WorkOrder {
	now := time.Now().UTC()
	woID := uuid.New()
	return &domain.WorkOrder{
		ID:         woID,
		CustomerID: uuid.New(),
		VehicleID:  uuid.New(),
		RepairType: "commercial",
		Status:     "draft",
		Complaint:  "стук при повороте",
		MileageKm:  50000,
		LaborCost:  "0.00",
		PartsCost:  "0.00",
		TotalCost:  "0.00",
		Notes:      "test work order",
		CreatedAt:  now,
		UpdatedAt:  now,
		Labor: []domain.WorkOrderLabor{
			{
				ID:          uuid.New(),
				WorkOrderID: woID,
				Description: "Замена колодок",
				Quantity:    "1.0",
				UnitPrice:   "2000.00",
				Amount:      "2000.00",
				SortOrder:   1,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		Parts: []domain.WorkOrderPart{
			{
				ID:          uuid.New(),
				WorkOrderID: woID,
				PartID:      uuid.New(),
				WarehouseID: uuid.New(),
				Description: "Колодка тормозная",
				Quantity:    "2.0",
				UnitPrice:   "1500.00",
				Amount:      "3000.00",
				SortOrder:   1,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
	}
}

func TestWorkOrderRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewWorkOrderRepository(testPool)

	wo := newTestWorkOrder()
	num, err := repo.NextOrderNumber(ctx)
	if err != nil {
		t.Fatalf("NextOrderNumber: %v", err)
	}
	wo.OrderNumber = num
	if err := repo.Create(ctx, wo); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, wo.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.OrderNumber != num || len(got.Labor) != 1 || len(got.Parts) != 1 {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
	if got.Labor[0].Amount != "2000.00" || got.Parts[0].Amount != "3000.00" {
		t.Fatalf("lines mismatch: %+v %+v", got.Labor[0], got.Parts[0])
	}

	openedAt := time.Now().UTC()
	wo.Status = "in_progress"
	wo.OpenedAt = &openedAt
	wo.MileageKm = 50100
	wo.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, wo, false); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ = repo.GetByID(ctx, wo.ID)
	if got.Status != "in_progress" || got.OpenedAt == nil || got.MileageKm != 50100 {
		t.Fatalf("update not persisted: %+v", got)
	}

	if err := repo.Delete(ctx, wo.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, wo.ID); err != pgx.ErrNoRows {
		t.Fatalf("expected ErrNoRows after delete, got %v", err)
	}
}

func TestWorkOrderRepository_ReplaceLines(t *testing.T) {
	ctx := context.Background()
	repo := NewWorkOrderRepository(testPool)

	wo := newTestWorkOrder()
	num, _ := repo.NextOrderNumber(ctx)
	wo.OrderNumber = num
	if err := repo.Create(ctx, wo); err != nil {
		t.Fatalf("Create: %v", err)
	}

	now := time.Now().UTC()
	wo.Labor = []domain.WorkOrderLabor{
		{
			ID:          uuid.New(),
			WorkOrderID: wo.ID,
			Description: "Диагностика",
			Quantity:    "1.0",
			UnitPrice:   "500.00",
			Amount:      "500.00",
			SortOrder:   1,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	wo.Parts = []domain.WorkOrderPart{
		{
			ID:          uuid.New(),
			WorkOrderID: wo.ID,
			PartID:      uuid.New(),
			WarehouseID: uuid.New(),
			Description: "Новая запчасть",
			Quantity:    "1.0",
			UnitPrice:   "700.00",
			Amount:      "700.00",
			SortOrder:   1,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	wo.UpdatedAt = now
	if err := repo.Update(ctx, wo, true); err != nil {
		t.Fatalf("Update replaceLines: %v", err)
	}
	got, _ := repo.GetByID(ctx, wo.ID)
	if len(got.Labor) != 1 || got.Labor[0].Description != "Диагностика" {
		t.Fatalf("labor replace failed: %+v", got.Labor)
	}
	if len(got.Parts) != 1 || got.Parts[0].Description != "Новая запчасть" {
		t.Fatalf("parts replace failed: %+v", got.Parts)
	}
}

func TestWorkOrderRepository_MarkPartsIssued(t *testing.T) {
	ctx := context.Background()
	repo := NewWorkOrderRepository(testPool)

	wo := newTestWorkOrder()
	num, _ := repo.NextOrderNumber(ctx)
	wo.OrderNumber = num
	if err := repo.Create(ctx, wo); err != nil {
		t.Fatalf("Create: %v", err)
	}

	docID := uuid.New()
	if err := repo.SetMovementDocument(ctx, wo.ID, docID, "draft"); err != nil {
		t.Fatalf("SetMovementDocument: %v", err)
	}
	got, _ := repo.GetByID(ctx, wo.ID)
	if got.MovementDocumentID == nil || *got.MovementDocumentID != docID || got.MovementDocumentStatus != "draft" {
		t.Fatalf("movement document not persisted: %+v", got)
	}

	lineID := wo.Parts[0].ID
	issuedAt := time.Now().UTC()
	if err := repo.MarkPartsIssued(ctx, wo.ID, []uuid.UUID{lineID}, issuedAt); err != nil {
		t.Fatalf("MarkPartsIssued: %v", err)
	}
	got, _ = repo.GetByID(ctx, wo.ID)
	if !got.PartsIssued || got.PartsIssuedAt == nil {
		t.Fatalf("parts_issued not persisted: %+v", got)
	}
	if got.MovementDocumentStatus != "closed" {
		t.Fatalf("movement_document_status should be closed, got %q", got.MovementDocumentStatus)
	}
	if !got.Parts[0].Issued {
		t.Fatal("part line should be issued")
	}
}

func TestWorkOrderRepository_ListFilters(t *testing.T) {
	ctx := context.Background()
	repo := NewWorkOrderRepository(testPool)

	customer := uuid.New()
	wo1 := newTestWorkOrder()
	wo1.CustomerID = customer
	wo1.RepairType = "maintenance"
	num1, _ := repo.NextOrderNumber(ctx)
	wo1.OrderNumber = num1
	wo1.Status = "in_progress"
	wo2 := newTestWorkOrder()
	wo2.RepairType = "maintenance"
	num2, _ := repo.NextOrderNumber(ctx)
	wo2.OrderNumber = num2
	wo2.Status = "closed"
	for _, wo := range []*domain.WorkOrder{wo1, wo2} {
		if err := repo.Create(ctx, wo); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	list, total, err := repo.List(ctx, 10, 0, "in_progress", "", "", "")
	if err != nil {
		t.Fatalf("List by status: %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].Status != "in_progress" {
		t.Fatalf("List by status: total=%d len=%d", total, len(list))
	}

	list, total, err = repo.List(ctx, 10, 0, "", "", customer.String(), "")
	if err != nil {
		t.Fatalf("List by customer: %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].ID != wo1.ID {
		t.Fatalf("List by customer: total=%d len=%d", total, len(list))
	}

	list, total, err = repo.List(ctx, 10, 0, "", "maintenance", "", "")
	if err != nil {
		t.Fatalf("List by repair type: %v", err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("List by repair type: total=%d len=%d", total, len(list))
	}
}

func TestWorkOrderRepository_NextOrderNumberMonotonic(t *testing.T) {
	ctx := context.Background()
	repo := NewWorkOrderRepository(testPool)

	n1, err := repo.NextOrderNumber(ctx)
	if err != nil {
		t.Fatalf("NextOrderNumber: %v", err)
	}
	n2, err := repo.NextOrderNumber(ctx)
	if err != nil {
		t.Fatalf("NextOrderNumber: %v", err)
	}
	if n1 == n2 {
		t.Fatal("order numbers should be unique")
	}
}
