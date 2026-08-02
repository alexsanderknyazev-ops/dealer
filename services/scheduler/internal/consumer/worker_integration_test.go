//go:build integration

package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"

	"github.com/dealer/dealer/pkg/appointmentevent"
	"github.com/dealer/dealer/pkg/dbschema"
	pkgkafka "github.com/dealer/dealer/pkg/kafka"
	"github.com/dealer/dealer/pkg/postgres"
	tc "github.com/dealer/dealer/pkg/testcontainers"
	"github.com/dealer/dealer/scheduler-service/internal/repository"
)

func TestWorker_CreatesNotificationForAppointmentEvent(t *testing.T) {
	ctx := context.Background()
	if !tc.DockerAvailable() {
		t.Skip("docker unavailable")
	}

	pg, err := tc.StartPostgres(ctx)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	t.Cleanup(func() { _ = pg.Close(ctx) })

	pool, err := postgres.NewPool(ctx, pg.DSN, dbschema.Reviews, dbschema.WorkOrders, dbschema.Deals, dbschema.Customers, dbschema.Clients, dbschema.Vehicles, dbschema.Appointments, dbschema.Public)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)

	kf, err := tc.StartKafka(ctx)
	if err != nil {
		t.Fatalf("start kafka: %v", err)
	}
	t.Cleanup(func() { _ = kf.Close(ctx) })

	topic := fmt.Sprintf("appointment-created-%d", time.Now().UnixNano())
	group := fmt.Sprintf("scheduler-int-%d", time.Now().UnixNano())

	client := &kafka.Client{Addr: kafka.TCP(kf.Brokers[0])}
	if _, err := client.CreateTopics(ctx, &kafka.CreateTopicsRequest{
		Topics: []kafka.TopicConfig{{Topic: topic, NumPartitions: 1, ReplicationFactor: 1}},
	}); err != nil {
		t.Fatalf("create topic: %v", err)
	}

	customerID := uuid.New()
	clientID := uuid.New()
	userID := uuid.New()
	email := uuid.NewString() + "@test.local"
	mustExec(t, pool, `INSERT INTO customers.customers (id, name, email, phone) VALUES ($1, '', $2, '')`, customerID, email)
	mustExec(t, pool, `INSERT INTO clients.clients (id, user_id, email, full_name, phone) VALUES ($1, $2, $3, '', '')`, clientID, userID, email)
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM clients.clients WHERE id = $1`, clientID)
		_, _ = pool.Exec(ctx, `DELETE FROM customers.customers WHERE id = $1`, customerID)
	})

	apptID := uuid.New()
	vehicleID := uuid.New()
	start := time.Now().UTC().Add(72 * time.Hour)
	mustExec(t, pool,
		`INSERT INTO appointments.repair_appointments (id, appointment_number, customer_id, vehicle_id, scheduled_start, scheduled_end, status) VALUES ($1, $2, $3, $4, $5, $6, 'scheduled')`,
		apptID, "RA-INT-0001", customerID, vehicleID, start, start.Add(2*time.Hour))
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM appointments.repair_appointments WHERE id = $1`, apptID)
	})

	producer := pkgkafka.NewProducer(kf.Brokers, topic)
	t.Cleanup(func() { _ = producer.Close() })
	ev := appointmentevent.CreatedEvent{
		Event:             appointmentevent.Created,
		AppointmentID:     apptID.String(),
		AppointmentNumber: "RA-INT-0001",
		CustomerID:        customerID.String(),
		VehicleID:         vehicleID.String(),
		ScheduledStart:    start.Unix(),
		ScheduledEnd:      start.Add(2 * time.Hour).Unix(),
		OccurredAt:        time.Now().Unix(),
	}
	body, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := producer.Publish(ctx, []byte(apptID.String()), body); err != nil {
		t.Fatalf("publish: %v", err)
	}

	kConsumer := pkgkafka.NewConsumer(kf.Brokers, topic, group)
	t.Cleanup(func() { _ = kConsumer.Close() })

	notifRepo := repository.NewNotificationRepository(pool)
	w := NewWorker(kConsumer, notifRepo, slog.New(discardHandler{}))
	go w.Run(ctx)

	deadline := time.Now().Add(90 * time.Second)
	for {
		var count int64
		_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM clients.client_notifications WHERE kind = 'repair_appointment_booked' AND source_id = $1`, apptID).Scan(&count)
		if count == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("notification not created in time, count = %d", count)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func mustExec(t *testing.T, pool *pgxpool.Pool, query string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), query, args...); err != nil {
		t.Fatalf("exec %s: %v", query, err)
	}
}

type discardHandler struct{}

func (discardHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (discardHandler) Handle(context.Context, slog.Record) error { return nil }
func (h discardHandler) WithAttrs([]slog.Attr) slog.Handler      { return h }
func (h discardHandler) WithGroup(string) slog.Handler           { return h }
