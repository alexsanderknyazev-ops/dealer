package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/client-registration-service/internal/domain"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func (r *NotificationRepository) ListUnreadByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.ClientNotification, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, client_id, user_id, kind, source_type, source_id, title, body, status, created_at, updated_at
		FROM client_notifications
		WHERE user_id = $1 AND status = 'unread'
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.ClientNotification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

func (r *NotificationRepository) Dismiss(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE client_notifications
		SET status = 'dismissed', updated_at = $3
		WHERE id = $1 AND user_id = $2 AND status = 'unread'
	`, id, userID, time.Now().UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func scanNotification(row pgx.Row) (*domain.ClientNotification, error) {
	var n domain.ClientNotification
	err := row.Scan(
		&n.ID, &n.ClientID, &n.UserID, &n.Kind, &n.SourceType, &n.SourceID,
		&n.Title, &n.Body, &n.Status, &n.CreatedAt, &n.UpdatedAt,
	)
	return &n, err
}
