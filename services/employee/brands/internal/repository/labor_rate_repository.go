package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/brands/internal/domain"
)

type LaborRateRepository struct {
	pool *pgxpool.Pool
}

func NewLaborRateRepository(pool *pgxpool.Pool) *LaborRateRepository {
	return &LaborRateRepository{pool: pool}
}

func scanRate(row pgx.Row) (*domain.BrandLaborRate, error) {
	var r domain.BrandLaborRate
	err := row.Scan(
		&r.ID, &r.BrandID, &r.DealerPointID,
		&r.WarrantyHourPrice, &r.CommercialHourPrice,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

const rateColumns = `id, brand_id, dealer_point_id, warranty_hour_price::text, commercial_hour_price::text, created_at, updated_at`

func (r *LaborRateRepository) GetByBrandAndDealerPoint(ctx context.Context, brandID, dealerPointID uuid.UUID) (*domain.BrandLaborRate, error) {
	return scanRate(r.pool.QueryRow(ctx,
		`SELECT `+rateColumns+` FROM brand_labor_rates WHERE brand_id = $1 AND dealer_point_id = $2`,
		brandID, dealerPointID,
	))
}

func (r *LaborRateRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.BrandLaborRate, error) {
	return scanRate(r.pool.QueryRow(ctx,
		`SELECT `+rateColumns+` FROM brand_labor_rates WHERE id = $1`, id,
	))
}

func (r *LaborRateRepository) List(ctx context.Context, limit, offset int32, brandID, dealerPointID *uuid.UUID) ([]*domain.BrandLaborRate, int32, error) {
	countQuery := `SELECT COUNT(*) FROM brand_labor_rates WHERE 1=1`
	listQuery := `SELECT ` + rateColumns + ` FROM brand_labor_rates WHERE 1=1`
	args := []interface{}{}
	argNum := 1
	if brandID != nil {
		countQuery += fmt.Sprintf(" AND brand_id = $%d", argNum)
		listQuery += fmt.Sprintf(" AND brand_id = $%d", argNum)
		args = append(args, *brandID)
		argNum++
	}
	if dealerPointID != nil {
		countQuery += fmt.Sprintf(" AND dealer_point_id = $%d", argNum)
		listQuery += fmt.Sprintf(" AND dealer_point_id = $%d", argNum)
		args = append(args, *dealerPointID)
		argNum++
	}
	var total int32
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listQuery += fmt.Sprintf(" ORDER BY brand_id, dealer_point_id LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*domain.BrandLaborRate
	for rows.Next() {
		item, err := scanRate(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}
	return list, total, nil
}

func (r *LaborRateRepository) Upsert(ctx context.Context, rate *domain.BrandLaborRate) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO brand_labor_rates (id, brand_id, dealer_point_id, warranty_hour_price, commercial_hour_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4::numeric, $5::numeric, $6, $7)
		ON CONFLICT (brand_id, dealer_point_id) DO UPDATE SET
			warranty_hour_price = EXCLUDED.warranty_hour_price,
			commercial_hour_price = EXCLUDED.commercial_hour_price,
			updated_at = EXCLUDED.updated_at`,
		rate.ID, rate.BrandID, rate.DealerPointID,
		rate.WarrantyHourPrice, rate.CommercialHourPrice,
		rate.CreatedAt, rate.UpdatedAt,
	)
	return err
}

func (r *LaborRateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM brand_labor_rates WHERE id = $1`, id)
	return err
}
