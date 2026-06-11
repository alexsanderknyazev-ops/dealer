package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/employee-statistics-service/internal/domain"
)

type StatsRepository struct {
	pool *pgxpool.Pool
}

func NewStatsRepository(pool *pgxpool.Pool) *StatsRepository {
	return &StatsRepository{pool: pool}
}

func (r *StatsRepository) InsertDealEvent(ctx context.Context, dealID, customerID, vehicleID uuid.UUID, amount, stage string, occurredAt time.Time) error {
	query := `
		INSERT INTO deal_events (deal_id, customer_id, vehicle_id, amount, stage, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (deal_id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, dealID, customerID, vehicleID, amount, stage, occurredAt)
	return err
}

func (r *StatsRepository) GetOverview(ctx context.Context) (*domain.Overview, error) {
	out := &domain.Overview{}

	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM deal_events`).Scan(&out.DealsCount); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT COALESCE(SUM(amount), 0) FROM deal_events`).Scan(&out.TotalRevenue); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT stage, COUNT(*)
		FROM deal_events
		GROUP BY stage
		ORDER BY stage
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.DealStageCount
		if err := rows.Scan(&item.Stage, &item.Count); err != nil {
			return nil, err
		}
		out.DealsByStage = append(out.DealsByStage, item)
	}
	return out, rows.Err()
}
