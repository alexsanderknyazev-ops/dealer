//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/client-reviews-service/internal/domain"
)

func seedClientAndVehicle(t *testing.T, ctx context.Context) (clientID, userID, vehicleID uuid.UUID) {
	t.Helper()
	clientID = uuid.New()
	userID = uuid.New()
	vehicleID = uuid.New()
	_, err := testPool.Exec(ctx,
		`INSERT INTO clients.clients (id, user_id, email, full_name, phone, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,now(),now())`,
		clientID, userID, "it.review."+clientID.String()+"@example.com", "Review Client", "+79990000030",
	)
	if err != nil {
		t.Fatalf("seed client: %v", err)
	}
	_, err = testPool.Exec(ctx,
		`INSERT INTO clients.client_vehicles (id, client_id, vehicle_id, vin, make, model, year, added_at) VALUES ($1,$2,$3,$4,$5,$6,$7,now())`,
		uuid.New(), clientID, vehicleID, "ITVINREVIEW"+clientID.String(), "Toyota", "Camry", 2021,
	)
	if err != nil {
		t.Fatalf("seed vehicle: %v", err)
	}
	return clientID, userID, vehicleID
}

func TestReviewRepository_ClientProfileAndVehicle(t *testing.T) {
	repo := NewReviewRepository(testPool)
	ctx := context.Background()

	clientID, userID, vehicleID := seedClientAndVehicle(t, ctx)

	email, fullName, err := repo.ClientProfile(ctx, clientID)
	if err != nil {
		t.Fatalf("ClientProfile: %v", err)
	}
	if email == "" || fullName != "Review Client" {
		t.Fatalf("ClientProfile: got email=%q fullName=%q", email, fullName)
	}

	gotClientID, err := repo.ClientVehicle(ctx, userID, vehicleID)
	if err != nil {
		t.Fatalf("ClientVehicle: %v", err)
	}
	if gotClientID != clientID {
		t.Fatalf("ClientVehicle: got %v want %v", gotClientID, clientID)
	}

	if _, err := repo.ClientVehicle(ctx, uuid.New(), vehicleID); !errors.Is(err, ErrNotOwner) {
		t.Fatalf("ClientVehicle wrong user: got %v want ErrNotOwner", err)
	}
}

func TestReviewRepository_CreateListGet(t *testing.T) {
	repo := NewReviewRepository(testPool)
	ctx := context.Background()

	clientID, userID, vehicleID := seedClientAndVehicle(t, ctx)
	now := time.Now().UTC()

	rv := &domain.Review{
		ID: uuid.New(), ClientID: clientID, UserID: userID, DealerPointID: uuid.New(),
		VehicleID: vehicleID, Rating: 5, Text: "Great service", Status: "published",
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Create(ctx, rv); err != nil {
		t.Fatalf("Create: %v", err)
	}

	list, err := repo.ListByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("list len: got %d want 1", len(list))
	}
	if list[0].ID != rv.ID || list[0].Rating != 5 || list[0].Text != rv.Text || list[0].Status != "published" {
		t.Fatalf("review mismatch: %+v", list[0])
	}

	got, err := repo.GetByIDForUser(ctx, rv.ID, userID)
	if err != nil {
		t.Fatalf("GetByIDForUser: %v", err)
	}
	if got.ID != rv.ID || got.ClientID != clientID || got.Status != "published" {
		t.Fatalf("GetByIDForUser mismatch: %+v", got)
	}

	if _, err := repo.GetByIDForUser(ctx, rv.ID, uuid.New()); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("GetByIDForUser wrong user: got %v want pgx.ErrNoRows", err)
	}
	if _, err := repo.GetByIDForUser(ctx, uuid.New(), userID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("GetByIDForUser missing: got %v want pgx.ErrNoRows", err)
	}
}

func TestReviewRepository_DuplicateReview(t *testing.T) {
	repo := NewReviewRepository(testPool)
	ctx := context.Background()

	clientID, userID, vehicleID := seedClientAndVehicle(t, ctx)
	now := time.Now().UTC()

	first := &domain.Review{
		ID: uuid.New(), ClientID: clientID, UserID: userID, DealerPointID: uuid.New(),
		VehicleID: vehicleID, Rating: 4, Text: "first", Status: "published", CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	dup := *first
	dup.ID = uuid.New()
	dup.Text = "second"
	dup.CreatedAt = now.Add(time.Hour)
	dup.UpdatedAt = now.Add(time.Hour)
	err := repo.Create(ctx, &dup)
	if !IsDuplicateReview(err) {
		t.Fatalf("duplicate Create: got %v want IsDuplicateReview", err)
	}
}

func TestReviewRepository_DismissInvitationForUser(t *testing.T) {
	repo := NewReviewRepository(testPool)
	ctx := context.Background()

	clientID, userID, vehicleID := seedClientAndVehicle(t, ctx)

	inv := &domain.ReviewInvitation{
		ID: uuid.New(), ClientID: clientID, UserID: userID, VehicleID: vehicleID,
		DealerPointID: uuid.New(), SourceType: "work_order", SourceID: uuid.New(),
		ServiceKind: "service", Status: "pending", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	_, err := testPool.Exec(ctx,
		`INSERT INTO review_invitations (id, client_id, user_id, vehicle_id, dealer_point_id, source_type, source_id, service_kind, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		inv.ID, inv.ClientID, inv.UserID, inv.VehicleID, inv.DealerPointID, inv.SourceType, inv.SourceID, inv.ServiceKind, inv.Status, inv.CreatedAt, inv.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("seed invitation: %v", err)
	}

	list, err := repo.ListPendingInvitationsByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("ListPendingInvitationsByUserID: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("pending len: got %d want 1", len(list))
	}
	if list[0].ID != inv.ID || list[0].Status != "pending" {
		t.Fatalf("invitation mismatch: %+v", list[0])
	}

	if err := repo.DismissInvitationForUser(ctx, inv.ID, userID); err != nil {
		t.Fatalf("DismissInvitationForUser: %v", err)
	}
	if err := repo.DismissInvitationForUser(ctx, inv.ID, userID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("second DismissInvitationForUser: got %v want pgx.ErrNoRows", err)
	}
	if err := repo.DismissInvitationForUser(ctx, inv.ID, uuid.New()); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("DismissInvitationForUser wrong user: got %v want pgx.ErrNoRows", err)
	}
}

func TestReviewRepository_CompleteInvitationsForVehicle(t *testing.T) {
	repo := NewReviewRepository(testPool)
	ctx := context.Background()

	clientID, userID, vehicleID := seedClientAndVehicle(t, ctx)

	insert := func(status string) uuid.UUID {
		t.Helper()
		id := uuid.New()
		_, err := testPool.Exec(ctx,
			`INSERT INTO review_invitations (id, client_id, user_id, vehicle_id, dealer_point_id, source_type, source_id, service_kind, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			id, clientID, userID, vehicleID, uuid.New(), "deal", uuid.New(), "sale", status, time.Now().UTC(), time.Now().UTC(),
		)
		if err != nil {
			t.Fatalf("seed invitation %s: %v", status, err)
		}
		return id
	}

	pendingID := insert("pending")
	_ = insert("completed")

	if err := repo.CompleteInvitationsForVehicle(ctx, clientID, vehicleID); err != nil {
		t.Fatalf("CompleteInvitationsForVehicle: %v", err)
	}

	var status string
	if err := testPool.QueryRow(ctx, `SELECT status FROM review_invitations WHERE id = $1`, pendingID).Scan(&status); err != nil {
		t.Fatalf("select status: %v", err)
	}
	if status != "completed" {
		t.Fatalf("status after complete: got %q want completed", status)
	}

	pending, err := repo.ListPendingInvitationsByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("ListPendingInvitationsByUserID: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("pending after complete: got %d want 0", len(pending))
	}
}
