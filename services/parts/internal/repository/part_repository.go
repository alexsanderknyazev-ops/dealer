package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

type PartRepository struct {
	pool *pgxpool.Pool
}

func NewPartRepository(pool *pgxpool.Pool) *PartRepository {
	return &PartRepository{pool: pool}
}

func (r *PartRepository) Create(ctx context.Context, p *domain.Part) error {
	query := `
		INSERT INTO parts (id, sku, name, category, folder_id, brand_id, dealer_point_id, legal_entity_id, warehouse_id, quantity, unit, price, location, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12::numeric, $13, $14, $15, $16)
	`
	_, err := r.pool.Exec(ctx, query,
		p.ID, p.SKU, p.Name, p.Category, p.FolderID, p.BrandID, p.DealerPointID, p.LegalEntityID, p.WarehouseID, p.Quantity, p.Unit, p.Price, p.Location, p.Notes, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *PartRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Part, error) {
	query := `
		SELECT id, sku, name, category, folder_id, brand_id, dealer_point_id, legal_entity_id, warehouse_id, quantity, unit, price::text, location, notes, created_at, updated_at
		FROM parts WHERE id = $1
	`
	var p domain.Part
	var price string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.SKU, &p.Name, &p.Category, &p.FolderID, &p.BrandID, &p.DealerPointID, &p.LegalEntityID, &p.WarehouseID, &p.Quantity, &p.Unit, &price, &p.Location, &p.Notes, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.Price = price
	return &p, nil
}

func (r *PartRepository) List(ctx context.Context, f domain.PartListFilter) ([]*domain.Part, int32, error) {
	limit, offset := f.Limit, f.Offset
	search, categoryFilter := f.Search, f.CategoryFilter
	folderID, brandID, dealerPointID, legalEntityID, warehouseID := f.FolderID, f.BrandID, f.DealerPointID, f.LegalEntityID, f.WarehouseID
	searchPattern := "%" + search + "%"
	countQuery := "SELECT COUNT(*) FROM parts WHERE 1=1"
	listQuery := `
		SELECT id, sku, name, category, folder_id, brand_id, dealer_point_id, legal_entity_id, warehouse_id, quantity, unit, price::text, location, notes, created_at, updated_at
		FROM parts WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1
	if search != "" {
		countQuery += fmt.Sprintf(" AND (sku ILIKE $%d OR name ILIKE $%d)", argNum, argNum)
		listQuery += fmt.Sprintf(" AND (sku ILIKE $%d OR name ILIKE $%d)", argNum, argNum)
		args = append(args, searchPattern)
		argNum++
	}
	if categoryFilter != "" {
		countQuery += fmt.Sprintf(" AND category = $%d", argNum)
		listQuery += fmt.Sprintf(" AND category = $%d", argNum)
		args = append(args, categoryFilter)
		argNum++
	}
	if folderID != nil {
		countQuery += fmt.Sprintf(" AND folder_id = $%d", argNum)
		listQuery += fmt.Sprintf(" AND folder_id = $%d", argNum)
		args = append(args, *folderID)
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
		// Запчасть может быть на нескольких складах (part_stock) или привязана к одному (parts.warehouse_id)
		countQuery += fmt.Sprintf(" AND (parts.warehouse_id = $%d OR parts.id IN (SELECT part_id FROM part_stock WHERE warehouse_id = $%d AND quantity > 0))", argNum, argNum)
		listQuery += fmt.Sprintf(" AND (parts.warehouse_id = $%d OR parts.id IN (SELECT part_id FROM part_stock WHERE warehouse_id = $%d AND quantity > 0))", argNum, argNum)
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
	var list []*domain.Part
	for rows.Next() {
		var p domain.Part
		var price string
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Category, &p.FolderID, &p.BrandID, &p.DealerPointID, &p.LegalEntityID, &p.WarehouseID, &p.Quantity, &p.Unit, &price, &p.Location, &p.Notes, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		p.Price = price
		list = append(list, &p)
	}
	return list, total, nil
}

func (r *PartRepository) Update(ctx context.Context, p *domain.Part) error {
	query := `
		UPDATE parts SET sku=$2, name=$3, category=$4, folder_id=$5, brand_id=$6, dealer_point_id=$7, legal_entity_id=$8, warehouse_id=$9, quantity=$10, unit=$11, price=$12::numeric, location=$13, notes=$14, updated_at=$15
		WHERE id=$1
	`
	_, err := r.pool.Exec(ctx, query, p.ID, p.SKU, p.Name, p.Category, p.FolderID, p.BrandID, p.DealerPointID, p.LegalEntityID, p.WarehouseID, p.Quantity, p.Unit, p.Price, p.Location, p.Notes, p.UpdatedAt)
	return err
}

func (r *PartRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM parts WHERE id = $1", id)
	return err
}
