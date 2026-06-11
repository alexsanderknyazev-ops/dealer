package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/dealerpoints/internal/domain"
)

type DealerPointRepository struct {
	pool *pgxpool.Pool
}

func NewDealerPointRepository(pool *pgxpool.Pool) *DealerPointRepository {
	return &DealerPointRepository{pool: pool}
}

func (r *DealerPointRepository) Create(ctx context.Context, d *domain.DealerPoint) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO dealer_points (id, name, address, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
		d.ID, d.Name, d.Address, d.CreatedAt, d.UpdatedAt,
	)
	return err
}

func (r *DealerPointRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.DealerPoint, error) {
	var d domain.DealerPoint
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, address, created_at, updated_at FROM dealer_points WHERE id = $1`, id,
	).Scan(&d.ID, &d.Name, &d.Address, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DealerPointRepository) List(ctx context.Context, limit, offset int32, search string) ([]*domain.DealerPoint, int32, error) {
	searchPattern := "%" + search + "%"
	countQuery := "SELECT COUNT(*) FROM dealer_points WHERE 1=1"
	listQuery := `SELECT id, name, address, created_at, updated_at FROM dealer_points WHERE 1=1`
	args := []interface{}{}
	argNum := 1
	if search != "" {
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR address ILIKE $%d)", argNum, argNum)
		listQuery += fmt.Sprintf(" AND (name ILIKE $%d OR address ILIKE $%d)", argNum, argNum)
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
	var list []*domain.DealerPoint
	for rows.Next() {
		var d domain.DealerPoint
		if err := rows.Scan(&d.ID, &d.Name, &d.Address, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &d)
	}
	return list, total, nil
}

func (r *DealerPointRepository) Update(ctx context.Context, d *domain.DealerPoint) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE dealer_points SET name=$2, address=$3, updated_at=$4 WHERE id=$1`,
		d.ID, d.Name, d.Address, d.UpdatedAt,
	)
	return err
}

func (r *DealerPointRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM dealer_points WHERE id = $1`, id)
	return err
}
