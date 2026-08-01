//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/client-registration-service/internal/domain"
)

func seedTestClient(t *testing.T, ctx context.Context) (clientID, userID uuid.UUID) {
	t.Helper()
	clientID = uuid.New()
	userID = uuid.New()
	_, err := testPool.Exec(ctx,
		`INSERT INTO clients (id, user_id, email, full_name, phone, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,now(),now())`,
		clientID, userID, "it.notif."+clientID.String()+"@example.com", "Notif Client", "+79990000020",
	)
	if err != nil {
		t.Fatalf("seed client: %v", err)
	}
	return clientID, userID
}

func TestNotificationRepository_ListUnreadByUserID(t *testing.T) {
	repo := NewNotificationRepository(testPool)
	ctx := context.Background()

	clientID, userID := seedTestClient(t, ctx)

	now := time.Now().UTC()
	notifs := []*domain.ClientNotification{
		{
			ID: uuid.New(), ClientID: clientID, UserID: userID, Kind: "customer_order_receipt",
			SourceType: "supplier_order", SourceID: uuid.New(), Title: "N1", Body: "B1",
			Status: "unread", CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now,
		},
		{
			ID: uuid.New(), ClientID: clientID, UserID: userID, Kind: "customer_order_receipt",
			SourceType: "supplier_order", SourceID: uuid.New(), Title: "N2", Body: "B2",
			Status: "unread", CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now,
		},
		{
			ID: uuid.New(), ClientID: clientID, UserID: userID, Kind: "customer_order_receipt",
			SourceType: "supplier_order", SourceID: uuid.New(), Title: "N3", Body: "B3",
			Status: "dismissed", CreatedAt: now, UpdatedAt: now,
		},
	}
	for _, n := range notifs {
		_, err := testPool.Exec(ctx,
			`INSERT INTO client_notifications (id, client_id, user_id, kind, source_type, source_id, title, body, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			n.ID, n.ClientID, n.UserID, n.Kind, n.SourceType, n.SourceID, n.Title, n.Body, n.Status, n.CreatedAt, n.UpdatedAt,
		)
		if err != nil {
			t.Fatalf("seed notification %s: %v", n.Title, err)
		}
	}

	list, err := repo.ListUnreadByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("ListUnreadByUserID: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("unread len: got %d want 2", len(list))
	}
	if list[0].Title != "N2" {
		t.Fatalf("order: got first %q want N2 (created_at DESC)", list[0].Title)
	}
	if list[0].Status != "unread" {
		t.Fatalf("status: got %q want unread", list[0].Status)
	}

	empty, err := repo.ListUnreadByUserID(ctx, uuid.New())
	if err != nil {
		t.Fatalf("ListUnreadByUserID other user: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("other user unread len: got %d want 0", len(empty))
	}
}

func TestNotificationRepository_Dismiss(t *testing.T) {
	repo := NewNotificationRepository(testPool)
	ctx := context.Background()

	clientID, userID := seedTestClient(t, ctx)

	n := &domain.ClientNotification{
		ID: uuid.New(), ClientID: clientID, UserID: userID, Kind: "customer_order_receipt",
		SourceType: "supplier_order", SourceID: uuid.New(), Title: "Dismiss", Body: "B",
		Status: "unread", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	_, err := testPool.Exec(ctx,
		`INSERT INTO client_notifications (id, client_id, user_id, kind, source_type, source_id, title, body, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		n.ID, n.ClientID, n.UserID, n.Kind, n.SourceType, n.SourceID, n.Title, n.Body, n.Status, n.CreatedAt, n.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("seed notification: %v", err)
	}

	if err := repo.Dismiss(ctx, userID, n.ID); err != nil {
		t.Fatalf("first Dismiss: %v", err)
	}
	if err := repo.Dismiss(ctx, userID, n.ID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("second Dismiss: got %v want pgx.ErrNoRows", err)
	}
	if err := repo.Dismiss(ctx, uuid.New(), n.ID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("Dismiss wrong user: got %v want pgx.ErrNoRows", err)
	}
}
