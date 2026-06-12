package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/client-reviews-service/internal/domain"
)

var ErrNotOwner = errors.New("vehicle not linked to client")

type ReviewRepository struct {
	pool *pgxpool.Pool
}

func NewReviewRepository(pool *pgxpool.Pool) *ReviewRepository {
	return &ReviewRepository{pool: pool}
}

func (r *ReviewRepository) ClientProfile(ctx context.Context, clientID uuid.UUID) (email, fullName string, err error) {
	query := `SELECT email, full_name FROM clients WHERE id = $1`
	err = r.pool.QueryRow(ctx, query, clientID).Scan(&email, &fullName)
	return email, fullName, err
}

func (r *ReviewRepository) ClientVehicle(ctx context.Context, userID, vehicleID uuid.UUID) (clientID uuid.UUID, err error) {
	query := `
		SELECT c.id
		FROM clients c
		JOIN client_vehicles cv ON cv.client_id = c.id
		WHERE c.user_id = $1 AND cv.vehicle_id = $2
	`
	err = r.pool.QueryRow(ctx, query, userID, vehicleID).Scan(&clientID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrNotOwner
	}
	return clientID, err
}

func (r *ReviewRepository) Create(ctx context.Context, review *domain.Review) error {
	query := `
		INSERT INTO reviews (id, client_id, user_id, dealer_point_id, vehicle_id, rating, text, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		review.ID, review.ClientID, review.UserID, review.DealerPointID, review.VehicleID,
		review.Rating, review.Text, review.Status, review.CreatedAt, review.UpdatedAt,
	)
	return err
}

func (r *ReviewRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Review, error) {
	query := `
		SELECT id, client_id, user_id, dealer_point_id, vehicle_id, rating, text, status, created_at, updated_at
		FROM reviews
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReviews(rows)
}

func (r *ReviewRepository) GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (*domain.Review, error) {
	query := `
		SELECT id, client_id, user_id, dealer_point_id, vehicle_id, rating, text, status, created_at, updated_at
		FROM reviews
		WHERE id = $1 AND user_id = $2
	`
	var review domain.Review
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&review.ID, &review.ClientID, &review.UserID, &review.DealerPointID, &review.VehicleID,
		&review.Rating, &review.Text, &review.Status, &review.CreatedAt, &review.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pgx.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func scanReviews(rows pgx.Rows) ([]*domain.Review, error) {
	var out []*domain.Review
	for rows.Next() {
		var review domain.Review
		if err := rows.Scan(
			&review.ID, &review.ClientID, &review.UserID, &review.DealerPointID, &review.VehicleID,
			&review.Rating, &review.Text, &review.Status, &review.CreatedAt, &review.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, &review)
	}
	return out, rows.Err()
}

func isUniqueViolation(err error) bool {
	var pgErr interface{ SQLState() string }
	return errors.As(err, &pgErr) && pgErr.SQLState() == "23505"
}

func IsDuplicateReview(err error) bool {
	return isUniqueViolation(err)
}

func (r *ReviewRepository) ListPendingInvitationsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.ReviewInvitation, error) {
	query := `
		SELECT id, client_id, user_id, vehicle_id, dealer_point_id,
		       source_type, source_id, service_kind, status, created_at, updated_at
		FROM review_invitations
		WHERE user_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.ReviewInvitation
	for rows.Next() {
		var inv domain.ReviewInvitation
		if err := rows.Scan(
			&inv.ID, &inv.ClientID, &inv.UserID, &inv.VehicleID, &inv.DealerPointID,
			&inv.SourceType, &inv.SourceID, &inv.ServiceKind, &inv.Status, &inv.CreatedAt, &inv.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, &inv)
	}
	return out, rows.Err()
}

func (r *ReviewRepository) DismissInvitationForUser(ctx context.Context, id, userID uuid.UUID) error {
	query := `
		UPDATE review_invitations
		SET status = 'dismissed', updated_at = now()
		WHERE id = $1 AND user_id = $2 AND status = 'pending'
	`
	tag, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *ReviewRepository) CompleteInvitationsForVehicle(ctx context.Context, clientID, vehicleID uuid.UUID) error {
	query := `
		UPDATE review_invitations
		SET status = 'completed', updated_at = now()
		WHERE client_id = $1 AND vehicle_id = $2 AND status = 'pending'
	`
	_, err := r.pool.Exec(ctx, query, clientID, vehicleID)
	return err
}

func NowUTC() time.Time {
	return time.Now().UTC()
}
