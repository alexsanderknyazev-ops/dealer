package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/brands/internal/domain"
)

type BrandRepository struct {
	pool *pgxpool.Pool
}

func NewBrandRepository(pool *pgxpool.Pool) *BrandRepository {
	return &BrandRepository{pool: pool}
}

func (r *BrandRepository) Create(ctx context.Context, b *domain.Brand) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO brands (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
		b.ID, b.Name, b.CreatedAt, b.UpdatedAt,
	)
	return err
}

func (r *BrandRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Brand, error) {
	var b domain.Brand
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, created_at, updated_at FROM brands WHERE id = $1`, id,
	).Scan(&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BrandRepository) List(ctx context.Context, limit, offset int32, search string) ([]*domain.Brand, int32, error) {
	searchPattern := "%" + search + "%"
	countQuery := "SELECT COUNT(*) FROM brands WHERE 1=1"
	listQuery := `SELECT id, name, created_at, updated_at FROM brands WHERE 1=1`
	args := []interface{}{}
	argNum := 1
	if search != "" {
		countQuery += fmt.Sprintf(" AND name ILIKE $%d", argNum)
		listQuery += fmt.Sprintf(" AND name ILIKE $%d", argNum)
		args = append(args, searchPattern)
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
	var list []*domain.Brand
	for rows.Next() {
		var b domain.Brand
		if err := rows.Scan(&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &b)
	}
	return list, total, nil
}

func (r *BrandRepository) Update(ctx context.Context, b *domain.Brand) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE brands SET name=$2, updated_at=$3 WHERE id=$1`,
		b.ID, b.Name, b.UpdatedAt,
	)
	return err
}

func (r *BrandRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM brands WHERE id = $1`, id)
	return err
}
