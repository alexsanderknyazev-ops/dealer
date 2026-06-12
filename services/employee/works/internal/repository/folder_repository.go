package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/works/internal/domain"
)

type FolderRepository struct {
	pool *pgxpool.Pool
}

func NewFolderRepository(pool *pgxpool.Pool) *FolderRepository {
	return &FolderRepository{pool: pool}
}

func (r *FolderRepository) Create(ctx context.Context, f *domain.WorkFolder) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO work_folders (id, name, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, f.ID, f.Name, f.ParentID, f.CreatedAt, f.UpdatedAt)
	return err
}

func (r *FolderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkFolder, error) {
	var f domain.WorkFolder
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, parent_id, created_at, updated_at FROM work_folders WHERE id = $1
	`, id).Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FolderRepository) ListByParent(ctx context.Context, parentID *uuid.UUID) ([]*domain.WorkFolder, error) {
	var query string
	var args []any
	if parentID == nil {
		query = `SELECT id, name, parent_id, created_at, updated_at FROM work_folders WHERE parent_id IS NULL ORDER BY name`
	} else {
		query = `SELECT id, name, parent_id, created_at, updated_at FROM work_folders WHERE parent_id = $1 ORDER BY name`
		args = []any{*parentID}
	}
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.WorkFolder
	for rows.Next() {
		var f domain.WorkFolder
		if err := rows.Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &f)
	}
	return list, nil
}

func (r *FolderRepository) Update(ctx context.Context, f *domain.WorkFolder) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE work_folders SET name=$2, parent_id=$3, updated_at=$4 WHERE id=$1
	`, f.ID, f.Name, f.ParentID, f.UpdatedAt)
	return err
}

func (r *FolderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM work_folders WHERE id = $1`, id)
	return err
}
