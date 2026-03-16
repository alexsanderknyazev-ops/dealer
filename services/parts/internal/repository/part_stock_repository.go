package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

type PartStockRepository struct {
	pool *pgxpool.Pool
}

func NewPartStockRepository(pool *pgxpool.Pool) *PartStockRepository {
	return &PartStockRepository{pool: pool}
}

func (r *PartStockRepository) ListByPart(ctx context.Context, partID uuid.UUID) ([]*domain.PartStock, error) {
	query := `
		SELECT part_id, warehouse_id, quantity, created_at, updated_at
		FROM part_stock WHERE part_id = $1 ORDER BY warehouse_id
	`
	rows, err := r.pool.Query(ctx, query, partID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.PartStock
	for rows.Next() {
		var s domain.PartStock
		if err := rows.Scan(&s.PartID, &s.WarehouseID, &s.Quantity, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, nil
}

func (r *PartStockRepository) Upsert(ctx context.Context, partID, warehouseID uuid.UUID, quantity int32) error {
	query := `
		INSERT INTO part_stock (part_id, warehouse_id, quantity, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (part_id, warehouse_id) DO UPDATE SET quantity = $3, updated_at = now()
	`
	_, err := r.pool.Exec(ctx, query, partID, warehouseID, quantity)
	return err
}

func (r *PartStockRepository) Delete(ctx context.Context, partID, warehouseID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM part_stock WHERE part_id = $1 AND warehouse_id = $2", partID, warehouseID)
	return err
}

func (r *PartStockRepository) ReplaceForPart(ctx context.Context, partID uuid.UUID, rows []struct{ WarehouseID uuid.UUID; Quantity int32 }) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, "DELETE FROM part_stock WHERE part_id = $1", partID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.Quantity < 0 {
			continue
		}
		_, err = tx.Exec(ctx,
			"INSERT INTO part_stock (part_id, warehouse_id, quantity, updated_at) VALUES ($1, $2, $3, now())",
			partID, row.WarehouseID, row.Quantity,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// HasStockInWarehouse — есть ли у запчасти остаток на этом складе (для фильтра списка)
func (r *PartStockRepository) PartIDsWithStockInWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT part_id FROM part_stock WHERE warehouse_id = $1 AND quantity > 0`
	rows, err := r.pool.Query(ctx, query, warehouseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
