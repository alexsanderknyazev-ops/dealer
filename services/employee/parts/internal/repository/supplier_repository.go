package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

type SupplierRepository struct {
	pool *pgxpool.Pool
}

func NewSupplierRepository(pool *pgxpool.Pool) *SupplierRepository {
	return &SupplierRepository{pool: pool}
}

func (r *SupplierRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM suppliers WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

func (r *SupplierRepository) Name(ctx context.Context, id uuid.UUID) string {
	var name string
	_ = r.pool.QueryRow(ctx, `SELECT name FROM suppliers WHERE id = $1`, id).Scan(&name)
	return name
}

func (r *SupplierRepository) List(ctx context.Context, limit, offset int32, search string) ([]*domain.Supplier, int32, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	countQuery := `SELECT COUNT(*) FROM suppliers WHERE 1=1`
	listQuery := `
		SELECT id, name, inn, phone, email, notes, created_at, updated_at
		FROM suppliers WHERE 1=1
	`
	args := []any{}
	argNum := 1
	if search != "" {
		countQuery += fmt.Sprintf(` AND (name ILIKE '%%' || $%d || '%%' OR inn ILIKE '%%' || $%d || '%%')`, argNum, argNum)
		listQuery += fmt.Sprintf(` AND (name ILIKE '%%' || $%d || '%%' OR inn ILIKE '%%' || $%d || '%%')`, argNum, argNum)
		args = append(args, search)
		argNum++
	}
	var total int32
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listQuery += fmt.Sprintf(` ORDER BY name LIMIT $%d OFFSET $%d`, argNum, argNum+1)
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []*domain.Supplier
	for rows.Next() {
		s, err := scanSupplier(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, s)
	}
	return out, total, nil
}

func scanSupplier(row pgx.Row) (*domain.Supplier, error) {
	var s domain.Supplier
	err := row.Scan(&s.ID, &s.Name, &s.INN, &s.Phone, &s.Email, &s.Notes, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}
