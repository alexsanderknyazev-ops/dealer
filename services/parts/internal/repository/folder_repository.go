package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

type FolderRepository struct {
	pool *pgxpool.Pool
}

func NewFolderRepository(pool *pgxpool.Pool) *FolderRepository {
	return &FolderRepository{pool: pool}
}

func (r *FolderRepository) Create(ctx context.Context, f *domain.PartFolder) error {
	query := `
		INSERT INTO part_folders (id, name, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query, f.ID, f.Name, f.ParentID, f.CreatedAt, f.UpdatedAt)
	return err
}

func (r *FolderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PartFolder, error) {
	query := `SELECT id, name, parent_id, created_at, updated_at FROM part_folders WHERE id = $1`
	var f domain.PartFolder
	err := r.pool.QueryRow(ctx, query, id).Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FolderRepository) ListByParent(ctx context.Context, parentID *uuid.UUID) ([]*domain.PartFolder, error) {
	var query string
	var args []interface{}
	if parentID == nil {
		query = `SELECT id, name, parent_id, created_at, updated_at FROM part_folders WHERE parent_id IS NULL ORDER BY name`
		args = []interface{}{}
	} else {
		query = `SELECT id, name, parent_id, created_at, updated_at FROM part_folders WHERE parent_id = $1 ORDER BY name`
		args = []interface{}{*parentID}
	}
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.PartFolder
	for rows.Next() {
		var f domain.PartFolder
		if err := rows.Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &f)
	}
	return list, nil
}

func (r *FolderRepository) Update(ctx context.Context, f *domain.PartFolder) error {
	query := `UPDATE part_folders SET name=$2, parent_id=$3, updated_at=$4 WHERE id=$1`
	_, err := r.pool.Exec(ctx, query, f.ID, f.Name, f.ParentID, f.UpdatedAt)
	return err
}

func (r *FolderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM part_folders WHERE id = $1", id)
	return err
}
