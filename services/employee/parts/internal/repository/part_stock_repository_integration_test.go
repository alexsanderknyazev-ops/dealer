//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

func TestPartStockRepository_UpsertAndQuantity(t *testing.T) {
	ctx := context.Background()
	partRepo := NewPartRepository(testPool)
	stockRepo := NewPartStockRepository(testPool)

	p := newTestPart()
	if err := partRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create part: %v", err)
	}
	wh := uuid.New()

	if err := stockRepo.Upsert(ctx, p.ID, wh, 7); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	qty, err := stockRepo.GetQuantity(ctx, p.ID, wh)
	if err != nil {
		t.Fatalf("GetQuantity: %v", err)
	}
	if qty != 7 {
		t.Fatalf("quantity mismatch: %d", qty)
	}

	if err := stockRepo.Upsert(ctx, p.ID, wh, 3); err != nil {
		t.Fatalf("Upsert overwrite: %v", err)
	}
	qty, _ = stockRepo.GetQuantity(ctx, p.ID, wh)
	if qty != 3 {
		t.Fatalf("upsert should overwrite, got %d", qty)
	}

	if err := stockRepo.Add(ctx, p.ID, wh, 10); err != nil {
		t.Fatalf("Add: %v", err)
	}
	qty, _ = stockRepo.GetQuantity(ctx, p.ID, wh)
	if qty != 13 {
		t.Fatalf("add should accumulate, got %d", qty)
	}

	remaining, err := stockRepo.Deduct(ctx, p.ID, wh, 5)
	if err != nil {
		t.Fatalf("Deduct: %v", err)
	}
	if remaining != 8 {
		t.Fatalf("deduct remaining mismatch: %d", remaining)
	}

	if _, err := stockRepo.Deduct(ctx, p.ID, wh, 100); err != pgx.ErrNoRows {
		t.Fatalf("expected ErrNoRows on insufficient stock, got %v", err)
	}

	if err := stockRepo.Delete(ctx, p.ID, wh); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := stockRepo.GetQuantity(ctx, p.ID, wh); err != pgx.ErrNoRows {
		t.Fatalf("expected ErrNoRows after delete, got %v", err)
	}
}

func TestPartStockRepository_ReplaceForPart(t *testing.T) {
	ctx := context.Background()
	partRepo := NewPartRepository(testPool)
	stockRepo := NewPartStockRepository(testPool)

	p := newTestPart()
	if err := partRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create part: %v", err)
	}
	wh1, wh2 := uuid.New(), uuid.New()
	if err := stockRepo.Add(ctx, p.ID, wh1, 4); err != nil {
		t.Fatalf("Add: %v", err)
	}

	rows := []domain.PartWarehouseQty{
		{WarehouseID: wh1, Quantity: 9},
		{WarehouseID: wh2, Quantity: 2},
	}
	if err := stockRepo.ReplaceForPart(ctx, p.ID, rows); err != nil {
		t.Fatalf("ReplaceForPart: %v", err)
	}

	list, err := stockRepo.ListByPart(ctx, p.ID)
	if err != nil {
		t.Fatalf("ListByPart: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 rows after replace, got %d", len(list))
	}
	qty1, _ := stockRepo.GetQuantity(ctx, p.ID, wh1)
	qty2, _ := stockRepo.GetQuantity(ctx, p.ID, wh2)
	if qty1 != 9 || qty2 != 2 {
		t.Fatalf("replace quantities mismatch: %d %d", qty1, qty2)
	}

	ids, err := stockRepo.PartIDsWithStockInWarehouse(ctx, wh2)
	if err != nil {
		t.Fatalf("PartIDsWithStockInWarehouse: %v", err)
	}
	found := false
	for _, id := range ids {
		if id == p.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("part should have stock in wh2, ids=%v", ids)
	}
}
