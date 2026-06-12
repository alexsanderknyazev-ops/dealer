package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/employee-reviews-service/internal/domain"
)

type ReviewRepository struct {
	pool *pgxpool.Pool
}

func NewReviewRepository(pool *pgxpool.Pool) *ReviewRepository {
	return &ReviewRepository{pool: pool}
}

const reviewSelect = `
	SELECT id, review_id, client_id, user_id, client_email, client_full_name,
	       dealer_point_id, vehicle_id, vehicle_vin, vehicle_make, vehicle_model, vehicle_year,
	       rating, text, status, occurred_at, created_at
	FROM reviews
`

func (r *ReviewRepository) Insert(ctx context.Context, review *domain.Review) error {
	query := `
		INSERT INTO reviews (
			review_id, client_id, user_id, client_email, client_full_name,
			dealer_point_id, vehicle_id, vehicle_vin, vehicle_make, vehicle_model, vehicle_year,
			rating, text, status, occurred_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		ON CONFLICT (review_id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query,
		review.ReviewID, review.ClientID, review.UserID, review.ClientEmail, review.ClientFullName,
		review.DealerPointID, review.VehicleID, review.VehicleVIN, review.VehicleMake, review.VehicleModel, review.VehicleYear,
		review.Rating, review.Text, review.Status, review.OccurredAt, review.CreatedAt,
	)
	return err
}

func (r *ReviewRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	query := reviewSelect + ` WHERE id = $1`
	var review domain.Review
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&review.ID, &review.ReviewID, &review.ClientID, &review.UserID, &review.ClientEmail, &review.ClientFullName,
		&review.DealerPointID, &review.VehicleID, &review.VehicleVIN, &review.VehicleMake, &review.VehicleModel, &review.VehicleYear,
		&review.Rating, &review.Text, &review.Status, &review.OccurredAt, &review.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *ReviewRepository) List(ctx context.Context, p domain.ReviewListParams) ([]*domain.Review, int64, error) {
	where, args := buildReviewFilters(p)
	limit := normalizeLimit(p.Limit)
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	countQuery := `SELECT COUNT(*) FROM reviews` + where
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listArgs := append(append([]any{}, args...), limit, offset)
	query := reviewSelect + where + ` ORDER BY occurred_at DESC LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)
	rows, err := r.pool.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list, err := scanReviews(rows)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *ReviewRepository) Stats(ctx context.Context, clientID, dealerPointID *uuid.UUID) (*domain.ReviewStats, error) {
	where, args := buildStatsFilters(clientID, dealerPointID)
	out := &domain.ReviewStats{}

	countQuery := `SELECT COUNT(*) FROM reviews` + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&out.TotalCount); err != nil {
		return nil, err
	}

	avgWhere := where
	if avgWhere == "" {
		avgWhere = ` WHERE status = 'published'`
	} else {
		avgWhere += ` AND status = 'published'`
	}
	avgQuery := `SELECT COALESCE(AVG(rating), 0) FROM reviews` + avgWhere
	if err := r.pool.QueryRow(ctx, avgQuery, args...).Scan(&out.AverageRating); err != nil {
		return nil, err
	}

	statusQuery := `SELECT status, COUNT(*) FROM reviews` + where + ` GROUP BY status ORDER BY status`
	rows, err := r.pool.Query(ctx, statusQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var sc domain.StatusCount
		if err := rows.Scan(&sc.Status, &sc.Count); err != nil {
			return nil, err
		}
		out.ByStatus = append(out.ByStatus, sc)
	}
	return out, rows.Err()
}

func buildReviewFilters(p domain.ReviewListParams) (string, []any) {
	var parts []string
	var args []any
	if p.ClientID != nil {
		args = append(args, *p.ClientID)
		parts = append(parts, fmt.Sprintf("client_id = $%d", len(args)))
	}
	if p.DealerPointID != nil {
		args = append(args, *p.DealerPointID)
		parts = append(parts, fmt.Sprintf("dealer_point_id = $%d", len(args)))
	}
	if p.Status != "" {
		args = append(args, p.Status)
		parts = append(parts, fmt.Sprintf("status = $%d", len(args)))
	}
	if len(parts) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}

func buildStatsFilters(clientID, dealerPointID *uuid.UUID) (string, []any) {
	p := domain.ReviewListParams{ClientID: clientID, DealerPointID: dealerPointID}
	return buildReviewFilters(p)
}

func normalizeLimit(limit int32) int32 {
	if limit <= 0 {
		return 50
	}
	if limit > 200 {
		return 200
	}
	return limit
}

type rowScanner interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanReviews(rows rowScanner) ([]*domain.Review, error) {
	var out []*domain.Review
	for rows.Next() {
		var review domain.Review
		if err := rows.Scan(
			&review.ID, &review.ReviewID, &review.ClientID, &review.UserID, &review.ClientEmail, &review.ClientFullName,
			&review.DealerPointID, &review.VehicleID, &review.VehicleVIN, &review.VehicleMake, &review.VehicleModel, &review.VehicleYear,
			&review.Rating, &review.Text, &review.Status, &review.OccurredAt, &review.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, &review)
	}
	return out, rows.Err()
}

func NowUTC() time.Time {
	return time.Now().UTC()
}
