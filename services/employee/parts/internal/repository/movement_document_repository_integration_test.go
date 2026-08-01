//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

func newTestMovementDocument(partID uuid.UUID) *domain.MovementDocument {
	now := time.Now().UTC()
	lineID := uuid.New()
	return &domain.MovementDocument{
		ID:           uuid.New(),
		Status:       "draft",
		MovementType: "receipt",
		Notes:        "test doc",
		CreatedAt:    now,
		UpdatedAt:    now,
		Lines: []domain.MovementDocumentLine{
			{
				ID:          lineID,
				PartID:      partID,
				WarehouseID: uuid.New(),
				Quantity:    5,
				UnitCost:    "1200.00",
				SortOrder:   1,
				CreatedAt:   now,
			},
		},
	}
}

func TestMovementDocumentRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	partRepo := NewPartRepository(testPool)
	repo := NewMovementDocumentRepository(testPool)

	p := newTestPart()
	if err := partRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create part: %v", err)
	}

	doc := newTestMovementDocument(p.ID)
	num, err := repo.NextDocumentNumber(ctx)
	if err != nil {
		t.Fatalf("NextDocumentNumber: %v", err)
	}
	doc.DocumentNumber = num
	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.DocumentNumber != num || got.MovementType != "receipt" || len(got.Lines) != 1 {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
	if got.Lines[0].PartID != p.ID || got.Lines[0].Quantity != 5 || got.Lines[0].UnitCost != "1200.00" {
		t.Fatalf("line mismatch: %+v", got.Lines[0])
	}

	confirmedBy := uuid.New()
	confirmedAt := time.Now().UTC()
	got.Status = "closed"
	got.ConfirmedBy = &confirmedBy
	got.ConfirmedAt = &confirmedAt
	got.UpdatedAt = time.Now().UTC()
	if err := repo.UpdateStatus(ctx, got); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	got2, err := repo.GetByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetByID after status: %v", err)
	}
	if got2.Status != "closed" || got2.ConfirmedBy == nil || *got2.ConfirmedBy != confirmedBy {
		t.Fatalf("status update not persisted: %+v", got2)
	}
}

func TestMovementDocumentRepository_ReplaceLines(t *testing.T) {
	ctx := context.Background()
	partRepo := NewPartRepository(testPool)
	repo := NewMovementDocumentRepository(testPool)

	p := newTestPart()
	if err := partRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create part: %v", err)
	}

	doc := newTestMovementDocument(p.ID)
	num, _ := repo.NextDocumentNumber(ctx)
	doc.DocumentNumber = num
	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create: %v", err)
	}

	now := time.Now().UTC()
	doc.Lines = []domain.MovementDocumentLine{
		{
			ID:          uuid.New(),
			PartID:      p.ID,
			WarehouseID: uuid.New(),
			Quantity:    9,
			UnitCost:    "900.00",
			SortOrder:   1,
			CreatedAt:   now,
		},
		{
			ID:          uuid.New(),
			PartID:      p.ID,
			WarehouseID: uuid.New(),
			Quantity:    1,
			UnitCost:    "800.00",
			SortOrder:   2,
			CreatedAt:   now,
		},
	}
	if err := repo.Update(ctx, doc, true); err != nil {
		t.Fatalf("Update replaceLines: %v", err)
	}
	got, err := repo.GetByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if len(got.Lines) != 2 {
		t.Fatalf("expected 2 lines after replace, got %d", len(got.Lines))
	}
}

func TestMovementDocumentRepository_ListFilters(t *testing.T) {
	ctx := context.Background()
	partRepo := NewPartRepository(testPool)
	repo := NewMovementDocumentRepository(testPool)

	p := newTestPart()
	if err := partRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create part: %v", err)
	}
	d1 := newTestMovementDocument(p.ID)
	num1, _ := repo.NextDocumentNumber(ctx)
	d1.DocumentNumber = num1
	d1.MovementType = "transfer"
	d1.ReferenceType = "work_order"
	if err := repo.Create(ctx, d1); err != nil {
		t.Fatalf("Create d1: %v", err)
	}
	d2 := newTestMovementDocument(p.ID)
	num2, _ := repo.NextDocumentNumber(ctx)
	d2.DocumentNumber = num2
	d2.MovementType = "receipt"
	d2.ReferenceType = "supplier_order"
	if err := repo.Create(ctx, d2); err != nil {
		t.Fatalf("Create d2: %v", err)
	}

	list, total, err := repo.List(ctx, 10, 0, "draft", "", "")
	if err != nil {
		t.Fatalf("List by status: %v", err)
	}
	if int(total) < 2 || len(list) < 2 {
		t.Fatalf("List by status draft: total=%d len=%d", total, len(list))
	}

	list, total, err = repo.List(ctx, 10, 0, "", "supplier_order", "")
	if err != nil {
		t.Fatalf("List by type: %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].MovementType != "receipt" {
		t.Fatalf("List by type: total=%d len=%d", total, len(list))
	}

	list, total, err = repo.List(ctx, 10, 0, "", "work_order", "")
	if err != nil {
		t.Fatalf("List by reference: %v", err)
	}
	if total < 1 {
		t.Fatalf("List by reference: total=%d", total)
	}
}

func TestMovementDocumentRepository_Extraction(t *testing.T) {
	ctx := context.Background()
	partRepo := NewPartRepository(testPool)
	repo := NewMovementDocumentRepository(testPool)

	p := newTestPart()
	if err := partRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create part: %v", err)
	}

	parent := newTestMovementDocument(p.ID)
	parent.MovementType = "to_production"
	parentNum, _ := repo.NextDocumentNumber(ctx)
	parent.DocumentNumber = parentNum
	if err := repo.Create(ctx, parent); err != nil {
		t.Fatalf("Create parent: %v", err)
	}
	parentLineID := parent.Lines[0].ID

	has, err := repo.HasOpenExtractionForParent(ctx, parent.ID)
	if err != nil {
		t.Fatalf("HasOpenExtractionForParent: %v", err)
	}
	if has {
		t.Fatal("no extraction yet, expected false")
	}

	child := newTestMovementDocument(p.ID)
	child.MovementType = "from_production"
	child.Status = "draft"
	child.ParentDocumentID = &parent.ID
	childNum, _ := repo.NextDocumentNumber(ctx)
	child.DocumentNumber = childNum
	child.Lines[0].ReferenceLineID = &parentLineID
	child.Lines[0].WarehouseID = parent.Lines[0].WarehouseID
	child.Lines[0].DestinationWarehouseID = &child.Lines[0].WarehouseID
	if err := repo.Create(ctx, child); err != nil {
		t.Fatalf("Create child: %v", err)
	}

	has, err = repo.HasOpenExtractionForParent(ctx, parent.ID)
	if err != nil {
		t.Fatalf("HasOpenExtractionForParent: %v", err)
	}
	if !has {
		t.Fatal("expected open extraction")
	}

	extracted, err := repo.ExtractedQuantityByParentLine(ctx, parent.ID, parentLineID)
	if err != nil {
		t.Fatalf("ExtractedQuantityByParentLine: %v", err)
	}
	if extracted != 0 {
		t.Fatalf("draft doc should not count, got %d", extracted)
	}

	confirmedBy := uuid.New()
	confirmedAt := time.Now().UTC()
	child.Status = "closed"
	child.ConfirmedBy = &confirmedBy
	child.ConfirmedAt = &confirmedAt
	child.UpdatedAt = time.Now().UTC()
	if err := repo.UpdateStatus(ctx, child); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	extracted, err = repo.ExtractedQuantityByParentLine(ctx, parent.ID, parentLineID)
	if err != nil {
		t.Fatalf("ExtractedQuantityByParentLine: %v", err)
	}
	if extracted != 5 {
		t.Fatalf("expected extracted 5, got %d", extracted)
	}
}

func TestMovementDocumentRepository_DocumentNumberUnique(t *testing.T) {
	ctx := context.Background()
	partRepo := NewPartRepository(testPool)
	repo := NewMovementDocumentRepository(testPool)

	p := newTestPart()
	if err := partRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create part: %v", err)
	}
	doc := newTestMovementDocument(p.ID)
	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create: %v", err)
	}
	dup := newTestMovementDocument(p.ID)
	dup.DocumentNumber = doc.DocumentNumber
	if err := repo.Create(ctx, dup); err == nil {
		t.Fatal("expected unique violation on duplicate document_number")
	}
}
