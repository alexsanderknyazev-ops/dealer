//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func notifMustExec(t *testing.T, ctx context.Context, query string, args ...any) {
	t.Helper()
	if _, err := testPool.Exec(ctx, query, args...); err != nil {
		t.Fatalf("exec %s: %v", query, err)
	}
}

func seedNotificationClient(t *testing.T) (customerID, clientID, userID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	email := uuid.NewString() + "@test.local"
	customerID, clientID, userID = uuid.New(), uuid.New(), uuid.New()
	notifMustExec(t, ctx, `INSERT INTO customers.customers (id, name, email, phone) VALUES ($1, '', $2, '')`, customerID, email)
	notifMustExec(t, ctx, `INSERT INTO clients.clients (id, user_id, email, full_name, phone) VALUES ($1, $2, $3, '', '')`, clientID, userID, email)
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM clients.clients WHERE id = $1`, clientID)
		_, _ = testPool.Exec(ctx, `DELETE FROM customers.customers WHERE id = $1`, customerID)
	})
	return customerID, clientID, userID
}

func TestNotificationRepository_CreateFromClosedCustomerOrderReceipts(t *testing.T) {
	ctx := context.Background()
	repo := NewNotificationRepository(testPool)
	customerID, _, _ := seedNotificationClient(t)

	supplierID := uuid.New()
	notifMustExec(t, ctx, `INSERT INTO parts.suppliers (id, name) VALUES ($1, '')`, supplierID)

	coID := uuid.New()
	notifMustExec(t, ctx, `INSERT INTO parts.customer_orders (id, order_number, status, customer_id, issue_warehouse_id) VALUES ($1, $2, 'fulfilled', $3, $4)`,
		coID, uuid.NewString(), customerID, uuid.New())
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM parts.customer_orders WHERE id = $1`, coID)
	})

	soID := uuid.New()
	notifMustExec(t, ctx, `INSERT INTO parts.supplier_orders (id, order_number, status, supplier_id, receipt_warehouse_id, customer_order_id) VALUES ($1, $2, 'fulfilled', $3, $4, $5)`,
		soID, uuid.NewString(), supplierID, uuid.New(), coID)
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM parts.supplier_orders WHERE id = $1`, soID)
	})

	mdID := uuid.New()
	notifMustExec(t, ctx, `INSERT INTO parts.movement_documents (id, document_number, status, movement_type, reference_type, reference_id) VALUES ($1, $2, 'closed', 'receipt', 'supplier_order', $3)`,
		mdID, uuid.NewString(), soID)
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM parts.movement_documents WHERE id = $1`, mdID)
	})

	n, err := repo.CreateFromClosedCustomerOrderReceipts(ctx, 10)
	if err != nil {
		t.Fatalf("CreateFromClosedCustomerOrderReceipts: %v", err)
	}
	if n < 1 {
		t.Fatalf("expected at least 1 notification, got %d", n)
	}

	var count int64
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM clients.client_notifications WHERE kind = 'customer_order_receipt' AND source_type = 'customer_order' AND source_id = $1`, coID).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 notification row, got %d", count)
	}

	n2, err := repo.CreateFromClosedCustomerOrderReceipts(ctx, 10)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if n2 != 0 {
		t.Fatalf("expected 0 new notifications on second call, got %d", n2)
	}
}

func TestNotificationRepository_CreateRepairAppointmentReminders(t *testing.T) {
	ctx := context.Background()
	repo := NewNotificationRepository(testPool)
	customerID, _, _ := seedNotificationClient(t)

	now := time.Now().UTC()
	dayAppt := uuid.New()
	dayStart := now.Add(24 * time.Hour)
	notifMustExec(t, ctx, `INSERT INTO appointments.repair_appointments (id, appointment_number, customer_id, vehicle_id, scheduled_start, scheduled_end, status) VALUES ($1, $2, $3, $4, $5, $6, 'scheduled')`,
		dayAppt, uuid.NewString(), customerID, uuid.New(), dayStart, dayStart.Add(2*time.Hour))

	hourAppt := uuid.New()
	hourStart := now.Add(1 * time.Hour)
	notifMustExec(t, ctx, `INSERT INTO appointments.repair_appointments (id, appointment_number, customer_id, vehicle_id, scheduled_start, scheduled_end, status) VALUES ($1, $2, $3, $4, $5, $6, 'scheduled')`,
		hourAppt, uuid.NewString(), customerID, uuid.New(), hourStart, hourStart.Add(2*time.Hour))

	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM appointments.repair_appointments WHERE id = $1`, dayAppt)
		_, _ = testPool.Exec(ctx, `DELETE FROM appointments.repair_appointments WHERE id = $1`, hourAppt)
	})

	n, err := repo.CreateRepairAppointmentReminders(ctx, 10)
	if err != nil {
		t.Fatalf("CreateRepairAppointmentReminders: %v", err)
	}
	if n < 2 {
		t.Fatalf("expected at least 2 reminders, got %d", n)
	}

	var dayCount, hourCount int64
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM clients.client_notifications WHERE kind = 'repair_appointment_reminder_24h' AND source_id = $1`, dayAppt).Scan(&dayCount); err != nil {
		t.Fatalf("day count: %v", err)
	}
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM clients.client_notifications WHERE kind = 'repair_appointment_reminder_1h' AND source_id = $1`, hourAppt).Scan(&hourCount); err != nil {
		t.Fatalf("hour count: %v", err)
	}
	if dayCount != 1 {
		t.Fatalf("expected 1 day reminder, got %d", dayCount)
	}
	if hourCount != 1 {
		t.Fatalf("expected 1 hour reminder, got %d", hourCount)
	}

	n2, err := repo.CreateRepairAppointmentReminders(ctx, 10)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if n2 != 0 {
		t.Fatalf("expected 0 new reminders on second call, got %d", n2)
	}
}
