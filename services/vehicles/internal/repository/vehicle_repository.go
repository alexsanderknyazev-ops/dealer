package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/vehicles/internal/domain"
)

type VehicleRepository struct {
	pool *pgxpool.Pool
}

func NewVehicleRepository(pool *pgxpool.Pool) *VehicleRepository {
	return &VehicleRepository{pool: pool}
}

func (r *VehicleRepository) Create(ctx context.Context, v *domain.Vehicle) error {
	query := `
		INSERT INTO vehicles (id, vin, make, model, year, mileage_km, price, status, color, notes, brand_id, dealer_point_id, legal_entity_id, warehouse_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7::numeric, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := r.pool.Exec(ctx, query,
		v.ID, v.VIN, v.Make, v.Model, v.Year, v.MileageKm, v.Price, v.Status, v.Color, v.Notes, v.BrandID, v.DealerPointID, v.LegalEntityID, v.WarehouseID, v.CreatedAt, v.UpdatedAt,
	)
	return err
}

func (r *VehicleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Vehicle, error) {
	query := `
		SELECT id, vin, make, model, year, mileage_km, price::text, status, color, notes, brand_id, dealer_point_id, legal_entity_id, warehouse_id, created_at, updated_at
		FROM vehicles WHERE id = $1
	`
	var v domain.Vehicle
	var price string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&v.ID, &v.VIN, &v.Make, &v.Model, &v.Year, &v.MileageKm, &price, &v.Status, &v.Color, &v.Notes, &v.BrandID, &v.DealerPointID, &v.LegalEntityID, &v.WarehouseID, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	v.Price = price
	return &v, nil
}

func (r *VehicleRepository) List(ctx context.Context, f domain.VehicleListFilter) ([]*domain.Vehicle, int32, error) {
	limit, offset := f.Limit, f.Offset
	search, statusFilter := f.Search, f.StatusFilter
	brandID, dealerPointID, legalEntityID, warehouseID := f.BrandID, f.DealerPointID, f.LegalEntityID, f.WarehouseID
	searchPattern := "%" + search + "%"
	countQuery := "SELECT COUNT(*) FROM vehicles WHERE 1=1"
	listQuery := `
		SELECT id, vin, make, model, year, mileage_km, price::text, status, color, notes, brand_id, dealer_point_id, legal_entity_id, warehouse_id, created_at, updated_at
		FROM vehicles WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1
	if search != "" {
		countQuery += fmt.Sprintf(" AND (vin ILIKE $%d OR make ILIKE $%d OR model ILIKE $%d)", argNum, argNum, argNum)
		listQuery += fmt.Sprintf(" AND (vin ILIKE $%d OR make ILIKE $%d OR model ILIKE $%d)", argNum, argNum, argNum)
		args = append(args, searchPattern)
		argNum++
	}
	if statusFilter != "" {
		countQuery += fmt.Sprintf(" AND status = $%d", argNum)
		listQuery += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, statusFilter)
		argNum++
	}
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
	if legalEntityID != nil {
		countQuery += fmt.Sprintf(" AND legal_entity_id = $%d", argNum)
		listQuery += fmt.Sprintf(" AND legal_entity_id = $%d", argNum)
		args = append(args, *legalEntityID)
		argNum++
	}
	if warehouseID != nil {
		countQuery += fmt.Sprintf(" AND warehouse_id = $%d", argNum)
		listQuery += fmt.Sprintf(" AND warehouse_id = $%d", argNum)
		args = append(args, *warehouseID)
		argNum++
	}
	var total int32
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listQuery += " ORDER BY created_at DESC LIMIT $" + fmt.Sprint(argNum) + " OFFSET $" + fmt.Sprint(argNum+1)
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*domain.Vehicle
	for rows.Next() {
		var v domain.Vehicle
		var price string
		if err := rows.Scan(&v.ID, &v.VIN, &v.Make, &v.Model, &v.Year, &v.MileageKm, &price, &v.Status, &v.Color, &v.Notes, &v.BrandID, &v.DealerPointID, &v.LegalEntityID, &v.WarehouseID, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, 0, err
		}
		v.Price = price
		list = append(list, &v)
	}
	return list, total, nil
}

func (r *VehicleRepository) Update(ctx context.Context, v *domain.Vehicle) error {
	query := `
		UPDATE vehicles SET vin=$2, make=$3, model=$4, year=$5, mileage_km=$6, price=$7::numeric, status=$8, color=$9, notes=$10, brand_id=$11, dealer_point_id=$12, legal_entity_id=$13, warehouse_id=$14, updated_at=$15
		WHERE id=$1
	`
	_, err := r.pool.Exec(ctx, query, v.ID, v.VIN, v.Make, v.Model, v.Year, v.MileageKm, v.Price, v.Status, v.Color, v.Notes, v.BrandID, v.DealerPointID, v.LegalEntityID, v.WarehouseID, v.UpdatedAt)
	return err
}

func (r *VehicleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM vehicles WHERE id = $1", id)
	return err
}
