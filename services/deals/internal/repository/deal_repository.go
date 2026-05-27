package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/deals/internal/domain"
)

type DealRepository struct {
	pool *pgxpool.Pool
}

func NewDealRepository(pool *pgxpool.Pool) *DealRepository {
	return &DealRepository{pool: pool}
}

func (r *DealRepository) Create(ctx context.Context, d *domain.Deal) error {
	query := `
		INSERT INTO deals (id, customer_id, vehicle_id, amount, stage, assigned_to, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4::numeric, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		d.ID, d.CustomerID, d.VehicleID, d.Amount, d.Stage, d.AssignedTo, d.Notes, d.CreatedAt, d.UpdatedAt,
	)
	return err
}

func (r *DealRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Deal, error) {
	query := `
		SELECT id, customer_id, vehicle_id, amount::text, stage, assigned_to, notes, created_at, updated_at
		FROM deals WHERE id = $1
	`
	var d domain.Deal
	var amount string
	var assignedTo *uuid.UUID
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.CustomerID, &d.VehicleID, &amount, &d.Stage, &assignedTo, &d.Notes, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	d.Amount = amount
	d.AssignedTo = assignedTo
	return &d, nil
}

func (r *DealRepository) List(ctx context.Context, limit, offset int32, stageFilter, customerID string) ([]*domain.Deal, int32, error) {
	countQuery := "SELECT COUNT(*) FROM deals WHERE 1=1"
	listQuery := `
		SELECT id, customer_id, vehicle_id, amount::text, stage, assigned_to, notes, created_at, updated_at
		FROM deals WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1
	if stageFilter != "" {
		countQuery += fmt.Sprintf(" AND stage = $%d", argNum)
		listQuery += fmt.Sprintf(" AND stage = $%d", argNum)
		args = append(args, stageFilter)
		argNum++
	}
	if customerID != "" {
		cid, err := uuid.Parse(customerID)
		if err == nil {
			countQuery += fmt.Sprintf(" AND customer_id = $%d", argNum)
			listQuery += fmt.Sprintf(" AND customer_id = $%d", argNum)
			args = append(args, cid)
			argNum++
		}
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
	var list []*domain.Deal
	for rows.Next() {
		var d domain.Deal
		var amount string
		var assignedTo *uuid.UUID
		if err := rows.Scan(&d.ID, &d.CustomerID, &d.VehicleID, &amount, &d.Stage, &assignedTo, &d.Notes, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, err
		}
		d.Amount = amount
		d.AssignedTo = assignedTo
		list = append(list, &d)
	}
	return list, total, nil
}

func (r *DealRepository) Update(ctx context.Context, d *domain.Deal) error {
	query := `
		UPDATE deals SET customer_id=$2, vehicle_id=$3, amount=$4::numeric, stage=$5, assigned_to=$6, notes=$7, updated_at=$8
		WHERE id=$1
	`
	_, err := r.pool.Exec(ctx, query, d.ID, d.CustomerID, d.VehicleID, d.Amount, d.Stage, d.AssignedTo, d.Notes, d.UpdatedAt)
	return err
}

func (r *DealRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM deals WHERE id = $1", id)
	return err
}
