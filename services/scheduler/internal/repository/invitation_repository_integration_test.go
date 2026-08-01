//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type invitationSeed struct {
	customerID    uuid.UUID
	vehicleID     uuid.UUID
	clientID      uuid.UUID
	userID        uuid.UUID
	dealerPointID uuid.UUID
	email         string
}

func invMustExec(t *testing.T, ctx context.Context, query string, args ...any) {
	t.Helper()
	if _, err := testPool.Exec(ctx, query, args...); err != nil {
		t.Fatalf("exec %s: %v", query, err)
	}
}

func seedInvitationSource(t *testing.T) invitationSeed {
	t.Helper()
	ctx := context.Background()
	s := invitationSeed{
		customerID:    uuid.New(),
		vehicleID:     uuid.New(),
		clientID:      uuid.New(),
		userID:        uuid.New(),
		dealerPointID: uuid.New(),
		email:         uuid.NewString() + "@test.local",
	}
	phone := "+7 900 000-00-00"
	invMustExec(t, ctx, `INSERT INTO dealerpoints.dealer_points (id, name, address) VALUES ($1, '', '')`, s.dealerPointID)
	invMustExec(t, ctx, `INSERT INTO customers.customers (id, name, email, phone) VALUES ($1, '', $2, $3)`, s.customerID, s.email, phone)
	invMustExec(t, ctx, `INSERT INTO vehicles.vehicles (id, vin, year, dealer_point_id) VALUES ($1, $2, 2020, $3)`, s.vehicleID, uuid.NewString(), s.dealerPointID)
	invMustExec(t, ctx, `INSERT INTO clients.clients (id, user_id, email, full_name, phone) VALUES ($1, $2, $3, '', $4)`, s.clientID, s.userID, s.email, phone)
	invMustExec(t, ctx, `INSERT INTO clients.client_vehicles (id, client_id, vehicle_id, vin) VALUES ($1, $2, $3, $4)`, uuid.New(), s.clientID, s.vehicleID, uuid.NewString())
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM clients.clients WHERE id = $1`, s.clientID)
		_, _ = testPool.Exec(ctx, `DELETE FROM customers.customers WHERE id = $1`, s.customerID)
		_, _ = testPool.Exec(ctx, `DELETE FROM vehicles.vehicles WHERE id = $1`, s.vehicleID)
	})
	return s
}

func TestInvitationRepository_CreateFromClosedWorkOrders(t *testing.T) {
	ctx := context.Background()
	repo := NewInvitationRepository(testPool)
	s := seedInvitationSource(t)

	woID := uuid.New()
	invMustExec(t, ctx, `INSERT INTO workorders.work_orders (id, order_number, customer_id, vehicle_id, dealer_point_id, status) VALUES ($1, $2, $3, $4, $5, 'closed')`,
		woID, uuid.NewString(), s.customerID, s.vehicleID, s.dealerPointID)
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM workorders.work_orders WHERE id = $1`, woID)
		_, _ = testPool.Exec(ctx, `DELETE FROM reviews.review_invitations WHERE source_type = 'work_order' AND source_id = $1`, woID)
	})

	n, err := repo.CreateFromClosedWorkOrders(ctx, 10)
	if err != nil {
		t.Fatalf("CreateFromClosedWorkOrders: %v", err)
	}
	if n < 1 {
		t.Fatalf("expected at least 1 invitation, got %d", n)
	}

	var count int64
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM reviews.review_invitations WHERE source_type = 'work_order' AND source_id = $1`, woID).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 invitation row for work order, got %d", count)
	}

	n2, err := repo.CreateFromClosedWorkOrders(ctx, 10)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if n2 != 0 {
		t.Fatalf("expected 0 new invitations on second call, got %d", n2)
	}
}

func TestInvitationRepository_CreateFromCompletedDeals(t *testing.T) {
	ctx := context.Background()
	repo := NewInvitationRepository(testPool)
	s := seedInvitationSource(t)

	dealID := uuid.New()
	invMustExec(t, ctx, `INSERT INTO deals.deals (id, customer_id, vehicle_id, stage) VALUES ($1, $2, $3, 'completed')`,
		dealID, s.customerID, s.vehicleID)
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM deals.deals WHERE id = $1`, dealID)
		_, _ = testPool.Exec(ctx, `DELETE FROM reviews.review_invitations WHERE source_type = 'deal' AND source_id = $1`, dealID)
	})

	n, err := repo.CreateFromCompletedDeals(ctx, 10)
	if err != nil {
		t.Fatalf("CreateFromCompletedDeals: %v", err)
	}
	if n < 1 {
		t.Fatalf("expected at least 1 invitation, got %d", n)
	}

	var count int64
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM reviews.review_invitations WHERE source_type = 'deal' AND source_id = $1`, dealID).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 invitation row for deal, got %d", count)
	}

	n2, err := repo.CreateFromCompletedDeals(ctx, 10)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if n2 != 0 {
		t.Fatalf("expected 0 new invitations on second call, got %d", n2)
	}
}

func TestInvitationRepository_CreateFromClosedGoodsSales(t *testing.T) {
	ctx := context.Background()
	repo := NewInvitationRepository(testPool)
	s := seedInvitationSource(t)

	legalID := uuid.New()
	invMustExec(t, ctx, `INSERT INTO dealerpoints.legal_entities (id, name, inn) VALUES ($1, '', $2)`, legalID, uuid.NewString())
	whID := uuid.New()
	invMustExec(t, ctx, `INSERT INTO dealerpoints.warehouses (id, dealer_point_id, legal_entity_id, type, name) VALUES ($1, $2, $3, 'parts', '')`,
		whID, s.dealerPointID, legalID)
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM dealerpoints.warehouses WHERE id = $1`, whID)
		_, _ = testPool.Exec(ctx, `DELETE FROM dealerpoints.legal_entities WHERE id = $1`, legalID)
	})

	partID := uuid.New()
	invMustExec(t, ctx, `INSERT INTO parts.parts (id, sku) VALUES ($1, $2)`, partID, uuid.NewString())
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM parts.parts WHERE id = $1`, partID)
	})

	mdID := uuid.New()
	invMustExec(t, ctx, `INSERT INTO parts.movement_documents (id, document_number, status, movement_type, customer_id, vehicle_id) VALUES ($1, $2, 'closed', 'sale', $3, $4)`,
		mdID, uuid.NewString(), s.customerID, s.vehicleID)
	invMustExec(t, ctx, `INSERT INTO parts.movement_document_lines (id, document_id, part_id, warehouse_id, quantity, sort_order) VALUES ($1, $2, $3, $4, 1, 0)`,
		uuid.New(), mdID, partID, whID)
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM parts.movement_documents WHERE id = $1`, mdID)
		_, _ = testPool.Exec(ctx, `DELETE FROM reviews.review_invitations WHERE source_type = 'movement_document' AND source_id = $1`, mdID)
	})

	n, err := repo.CreateFromClosedGoodsSales(ctx, 10)
	if err != nil {
		t.Fatalf("CreateFromClosedGoodsSales: %v", err)
	}
	if n < 1 {
		t.Fatalf("expected at least 1 invitation, got %d", n)
	}

	var count int64
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM reviews.review_invitations WHERE source_type = 'movement_document' AND source_id = $1`, mdID).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 invitation row for movement document, got %d", count)
	}

	n2, err := repo.CreateFromClosedGoodsSales(ctx, 10)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if n2 != 0 {
		t.Fatalf("expected 0 new invitations on second call, got %d", n2)
	}
}
