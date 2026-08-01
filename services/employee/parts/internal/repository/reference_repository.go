package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReferenceRepository struct {
	pool *pgxpool.Pool
}

func NewReferenceRepository(pool *pgxpool.Pool) *ReferenceRepository {
	return &ReferenceRepository{pool: pool}
}

func (r *ReferenceRepository) CustomerExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM customers.customers WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

func (r *ReferenceRepository) VehicleExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM vehicles.vehicles WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

func (r *ReferenceRepository) LookupVehicleIDByVIN(ctx context.Context, vin string) (*uuid.UUID, error) {
	vin = strings.TrimSpace(vin)
	if vin == "" {
		return nil, nil
	}
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id FROM vehicles.vehicles WHERE lower(trim(vin)) = lower(trim($1)) LIMIT 1
	`, vin).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func (r *ReferenceRepository) CustomerName(ctx context.Context, id uuid.UUID) string {
	var name string
	_ = r.pool.QueryRow(ctx, `SELECT name FROM customers.customers WHERE id = $1`, id).Scan(&name)
	return name
}

func (r *ReferenceRepository) VehicleInfo(ctx context.Context, id uuid.UUID) (vin, label string) {
	var make, model string
	var year int32
	err := r.pool.QueryRow(ctx, `
		SELECT vin, make, model, year FROM vehicles.vehicles WHERE id = $1
	`, id).Scan(&vin, &make, &model, &year)
	if err != nil {
		return "", ""
	}
	parts := []string{make, model}
	if year > 0 {
		parts = append(parts, fmt.Sprintf("%d", year))
	}
	return vin, strings.TrimSpace(strings.Join(parts, " "))
}
