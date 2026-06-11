package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/services/dealerpoints/internal/domain"
)

type LegalEntityRepository struct {
	pool *pgxpool.Pool
}

func NewLegalEntityRepository(pool *pgxpool.Pool) *LegalEntityRepository {
	return &LegalEntityRepository{pool: pool}
}

func (r *LegalEntityRepository) Create(ctx context.Context, e *domain.LegalEntity) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO legal_entities (id, name, inn, address, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		e.ID, e.Name, e.INN, e.Address, e.CreatedAt, e.UpdatedAt,
	)
	return err
}

func (r *LegalEntityRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.LegalEntity, error) {
	var e domain.LegalEntity
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, inn, address, created_at, updated_at FROM legal_entities WHERE id = $1`, id,
	).Scan(&e.ID, &e.Name, &e.INN, &e.Address, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *LegalEntityRepository) List(ctx context.Context, limit, offset int32, search string) ([]*domain.LegalEntity, int32, error) {
	searchPattern := "%" + search + "%"
	countQuery := "SELECT COUNT(*) FROM legal_entities WHERE 1=1"
	listQuery := `SELECT id, name, inn, address, created_at, updated_at FROM legal_entities WHERE 1=1`
	args := []interface{}{}
	argNum := 1
	if search != "" {
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR inn ILIKE $%d)", argNum, argNum)
		listQuery += fmt.Sprintf(" AND (name ILIKE $%d OR inn ILIKE $%d)", argNum, argNum)
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
	var list []*domain.LegalEntity
	for rows.Next() {
		var e domain.LegalEntity
		if err := rows.Scan(&e.ID, &e.Name, &e.INN, &e.Address, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &e)
	}
	return list, total, nil
}

func (r *LegalEntityRepository) ListByDealerPoint(ctx context.Context, dealerPointID uuid.UUID, limit, offset int32) ([]*domain.LegalEntity, int32, error) {
	countQuery := `SELECT COUNT(*) FROM dealer_point_legal_entities dple
		JOIN legal_entities le ON le.id = dple.legal_entity_id
		WHERE dple.dealer_point_id = $1`
	var total int32
	if err := r.pool.QueryRow(ctx, countQuery, dealerPointID).Scan(&total); err != nil {
		return nil, 0, err
	}
	listQuery := `SELECT le.id, le.name, le.inn, le.address, le.created_at, le.updated_at
		FROM dealer_point_legal_entities dple
		JOIN legal_entities le ON le.id = dple.legal_entity_id
		WHERE dple.dealer_point_id = $1
		ORDER BY le.name LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, listQuery, dealerPointID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*domain.LegalEntity
	for rows.Next() {
		var e domain.LegalEntity
		if err := rows.Scan(&e.ID, &e.Name, &e.INN, &e.Address, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &e)
	}
	return list, total, nil
}

func (r *LegalEntityRepository) Update(ctx context.Context, e *domain.LegalEntity) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE legal_entities SET name=$2, inn=$3, address=$4, updated_at=$5 WHERE id=$1`,
		e.ID, e.Name, e.INN, e.Address, e.UpdatedAt,
	)
	return err
}

func (r *LegalEntityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM legal_entities WHERE id = $1`, id)
	return err
}

func (r *LegalEntityRepository) LinkToDealerPoint(ctx context.Context, dealerPointID, legalEntityID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO dealer_point_legal_entities (dealer_point_id, legal_entity_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		dealerPointID, legalEntityID,
	)
	return err
}

func (r *LegalEntityRepository) UnlinkFromDealerPoint(ctx context.Context, dealerPointID, legalEntityID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM dealer_point_legal_entities WHERE dealer_point_id = $1 AND legal_entity_id = $2`,
		dealerPointID, legalEntityID,
	)
	return err
}
