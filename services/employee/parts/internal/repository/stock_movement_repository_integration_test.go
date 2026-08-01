//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

func TestStockMovementRepository_Create(t *testing.T) {
	ctx := context.Background()
	partRepo := NewPartRepository(testPool)
	repo := NewStockMovementRepository(testPool)

	p := newTestPart()
	if err := partRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create part: %v", err)
	}

	m := &domain.StockMovement{
		ID:            uuid.New(),
		PartID:        p.ID,
		WarehouseID:   uuid.New(),
		Quantity:      3,
		MovementType:  "receipt",
		ReferenceType: "test",
		Notes:         "поступление",
		CreatedAt:     time.Now().UTC(),
	}
	if err := repo.Create(ctx, m); err != nil {
		t.Fatalf("Create: %v", err)
	}

	var got struct {
		PartID       uuid.UUID
		WarehouseID  uuid.UUID
		Quantity     int32
		MovementType string
	}
	if err := testPool.QueryRow(ctx,
		"SELECT part_id, warehouse_id, quantity, movement_type FROM stock_movements WHERE id = $1",
		m.ID).Scan(&got.PartID, &got.WarehouseID, &got.Quantity, &got.MovementType); err != nil {
		t.Fatalf("fetch movement: %v", err)
	}
	if got.Quantity != 3 || got.MovementType != "receipt" {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
}

func TestStockMovementRepository_InvalidMovementType(t *testing.T) {
	ctx := context.Background()
	partRepo := NewPartRepository(testPool)
	repo := NewStockMovementRepository(testPool)

	p := newTestPart()
	if err := partRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create part: %v", err)
	}

	m := &domain.StockMovement{
		ID:           uuid.New(),
		PartID:       p.ID,
		WarehouseID:  uuid.New(),
		Quantity:     1,
		MovementType: "nope",
		CreatedAt:    time.Now().UTC(),
	}
	err := repo.Create(ctx, m)
	if err == nil {
		t.Fatal("expected CHECK violation for invalid movement_type")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23514" {
		t.Fatalf("expected SQLSTATE 23514, got %v", err)
	}
}
