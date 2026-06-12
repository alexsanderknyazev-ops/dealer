package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/works/internal/domain"
)

type WorkRepository struct {
	pool *pgxpool.Pool
}

func NewWorkRepository(pool *pgxpool.Pool) *WorkRepository {
	return &WorkRepository{pool: pool}
}

func (r *WorkRepository) Create(ctx context.Context, w *domain.Work) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO works (id, code, name, category, folder_id, labor_hours, unit_price, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6::numeric, $7::numeric, $8, $9, $10)
	`, w.ID, w.Code, w.Name, w.Category, w.FolderID, w.LaborHours, w.UnitPrice, w.Notes, w.CreatedAt, w.UpdatedAt)
	return err
}

func (r *WorkRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Work, error) {
	var w domain.Work
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, name, category, folder_id, labor_hours::text, unit_price::text, notes, created_at, updated_at
		FROM works WHERE id = $1
	`, id).Scan(&w.ID, &w.Code, &w.Name, &w.Category, &w.FolderID, &w.LaborHours, &w.UnitPrice, &w.Notes, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WorkRepository) List(ctx context.Context, limit, offset int32, search, category string, folderID *uuid.UUID) ([]*domain.Work, int32, error) {
	countQuery := "SELECT COUNT(*) FROM works WHERE 1=1"
	listQuery := `
		SELECT id, code, name, category, folder_id, labor_hours::text, unit_price::text, notes, created_at, updated_at
		FROM works WHERE 1=1
	`
	args := []any{}
	argNum := 1
	if search != "" {
		pattern := "%" + search + "%"
		clause := fmt.Sprintf(" AND (code ILIKE $%d OR name ILIKE $%d OR category ILIKE $%d)", argNum, argNum, argNum)
		countQuery += clause
		listQuery += clause
		args = append(args, pattern)
		argNum++
	}
	if category != "" {
		countQuery += fmt.Sprintf(" AND category = $%d", argNum)
		listQuery += fmt.Sprintf(" AND category = $%d", argNum)
		args = append(args, category)
		argNum++
	}
	if folderID != nil {
		countQuery += fmt.Sprintf(" AND folder_id = $%d", argNum)
		listQuery += fmt.Sprintf(" AND folder_id = $%d", argNum)
		args = append(args, *folderID)
		argNum++
	}
	var total int32
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listQuery += " ORDER BY category, name LIMIT $" + fmt.Sprint(argNum) + " OFFSET $" + fmt.Sprint(argNum+1)
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*domain.Work
	for rows.Next() {
		var w domain.Work
		if err := rows.Scan(&w.ID, &w.Code, &w.Name, &w.Category, &w.FolderID, &w.LaborHours, &w.UnitPrice, &w.Notes, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &w)
	}
	return list, total, nil
}

func (r *WorkRepository) Update(ctx context.Context, w *domain.Work) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE works
		SET code=$2, name=$3, category=$4, folder_id=$5, labor_hours=$6::numeric, unit_price=$7::numeric, notes=$8, updated_at=$9
		WHERE id=$1
	`, w.ID, w.Code, w.Name, w.Category, w.FolderID, w.LaborHours, w.UnitPrice, w.Notes, w.UpdatedAt)
	return err
}

func (r *WorkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM works WHERE id = $1`, id)
	return err
}
