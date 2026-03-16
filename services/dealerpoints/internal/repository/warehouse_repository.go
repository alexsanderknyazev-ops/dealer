package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/dealerpoints/internal/domain"
)

type WarehouseRepository struct {
	pool *pgxpool.Pool
}

func NewWarehouseRepository(pool *pgxpool.Pool) *WarehouseRepository {
	return &WarehouseRepository{pool: pool}
}

func (r *WarehouseRepository) Create(ctx context.Context, w *domain.Warehouse) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO warehouses (id, dealer_point_id, legal_entity_id, type, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		w.ID, w.DealerPointID, w.LegalEntityID, w.Type, w.Name, w.CreatedAt, w.UpdatedAt,
	)
	return err
}

func (r *WarehouseRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	var w domain.Warehouse
	err := r.pool.QueryRow(ctx,
		`SELECT id, dealer_point_id, legal_entity_id, type, name, created_at, updated_at FROM warehouses WHERE id = $1`, id,
	).Scan(&w.ID, &w.DealerPointID, &w.LegalEntityID, &w.Type, &w.Name, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WarehouseRepository) List(ctx context.Context, limit, offset int32, dealerPointID, legalEntityID *uuid.UUID, typeFilter string) ([]*domain.Warehouse, int32, error) {
	countQuery := "SELECT COUNT(*) FROM warehouses WHERE 1=1"
	listQuery := `SELECT id, dealer_point_id, legal_entity_id, type, name, created_at, updated_at FROM warehouses WHERE 1=1`
	args := []interface{}{}
	argNum := 1
	if dealerPointID != nil {
		countQuery += fmt.Sprintf(" AND dealer_point_id = $%d", argNum)
		listQuery += fmt.Sprintf(" AND dealer_point_id = $%d", argNum)
		args = append(args, *dealerPointID)
		argNum++
	}
	if legalEntityID != nil {
		countQuery += fmt.Sprintf(" AND legal_entity_id = $%d", argNum)
		listQuery += fmt.Sprintf(" AND legal_entity_id = $%d", argNum)
		args = append(args, *legalEntityID)
		argNum++
	}
	if typeFilter != "" {
		countQuery += fmt.Sprintf(" AND type = $%d", argNum)
		listQuery += fmt.Sprintf(" AND type = $%d", argNum)
		args = append(args, typeFilter)
		argNum++
	}
	var total int32
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listQuery += " ORDER BY name LIMIT $" + fmt.Sprint(argNum) + " OFFSET $" + fmt.Sprint(argNum+1)
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*domain.Warehouse
	for rows.Next() {
		var w domain.Warehouse
		if err := rows.Scan(&w.ID, &w.DealerPointID, &w.LegalEntityID, &w.Type, &w.Name, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &w)
	}
	return list, total, nil
}

func (r *WarehouseRepository) Update(ctx context.Context, w *domain.Warehouse) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE warehouses SET name=$2, updated_at=$3 WHERE id=$1`,
		w.ID, w.Name, w.UpdatedAt,
	)
	return err
}

func (r *WarehouseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM warehouses WHERE id = $1`, id)
	return err
}
