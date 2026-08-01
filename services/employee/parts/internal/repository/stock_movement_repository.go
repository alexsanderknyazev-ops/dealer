package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

type StockMovementRepository struct {
	pool *pgxpool.Pool
}

func NewStockMovementRepository(pool *pgxpool.Pool) *StockMovementRepository {
	return &StockMovementRepository{pool: pool}
}

func (r *StockMovementRepository) Create(ctx context.Context, m *domain.StockMovement) error {
	query := `
		INSERT INTO stock_movements (
			id, part_id, warehouse_id, quantity, movement_type,
			reference_type, reference_id, reference_line_id, movement_document_id, notes, created_by, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`
	_, err := r.pool.Exec(ctx, query,
		m.ID, m.PartID, m.WarehouseID, m.Quantity, m.MovementType,
		m.ReferenceType, m.ReferenceID, m.ReferenceLineID, m.MovementDocumentID, m.Notes, m.CreatedBy, m.CreatedAt,
	)
	return err
}
