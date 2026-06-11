package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/client-statistics-service/internal/domain"
)

type StatsRepository struct {
	pool *pgxpool.Pool
}

func NewStatsRepository(pool *pgxpool.Pool) *StatsRepository {
	return &StatsRepository{pool: pool}
}

func (r *StatsRepository) InsertClientRegistration(ctx context.Context, userID uuid.UUID, email string, vehicleID *uuid.UUID, occurredAt time.Time) error {
	query := `
		INSERT INTO client_registration_events (user_id, email, vehicle_id, occurred_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, userID, email, vehicleID, occurredAt)
	return err
}

func (r *StatsRepository) InsertReviewEvent(ctx context.Context, reviewID, clientID, userID, dealerPointID, vehicleID uuid.UUID, rating int32, status string, occurredAt time.Time) error {
	query := `
		INSERT INTO review_events (review_id, client_id, user_id, dealer_point_id, vehicle_id, rating, status, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (review_id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, reviewID, clientID, userID, dealerPointID, vehicleID, rating, status, occurredAt)
	return err
}

func (r *StatsRepository) GetOverview(ctx context.Context) (*domain.Overview, error) {
	out := &domain.Overview{}

	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM client_registration_events`).Scan(&out.ClientsCount); err != nil {
		return nil, err
	}
	out.RegisteredUsersCount = out.ClientsCount

	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT vehicle_id)
		FROM (
			SELECT vehicle_id FROM client_registration_events WHERE vehicle_id IS NOT NULL
			UNION
			SELECT vehicle_id FROM review_events
		) vehicles
	`).Scan(&out.ClientVehiclesCount); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM review_events`).Scan(&out.ReviewsCount); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(AVG(rating), 0)
		FROM review_events
		WHERE status = 'published'
	`).Scan(&out.AverageRating); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT status, COUNT(*)
		FROM review_events
		GROUP BY status
		ORDER BY status
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.ReviewStatusCount
		if err := rows.Scan(&item.Status, &item.Count); err != nil {
			return nil, err
		}
		out.ReviewsByStatus = append(out.ReviewsByStatus, item)
	}
	return out, rows.Err()
}
