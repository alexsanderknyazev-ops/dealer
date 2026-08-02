package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/appointments/internal/domain"
)

type fakeRepo struct {
	mu       sync.Mutex
	created  []*domain.RepairAppointment
	nextNum  string
	overlap  bool
}

func (f *fakeRepo) NextNumber(context.Context) (string, error) {
	return f.nextNum, nil
}

func (f *fakeRepo) Create(_ context.Context, a *domain.RepairAppointment) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.created = append(f.created, a)
	return nil
}

func (f *fakeRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.RepairAppointment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, a := range f.created {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (f *fakeRepo) Update(context.Context, *domain.RepairAppointment, bool) error { return nil }

func (f *fakeRepo) List(context.Context, int32, int32, string, *time.Time, *time.Time) ([]*domain.RepairAppointment, int32, error) {
	return nil, 0, nil
}

func (f *fakeRepo) HasOverlap(context.Context, time.Time, time.Time, *uuid.UUID) (bool, error) {
	return f.overlap, nil
}

func (f *fakeRepo) ListBusyInRange(context.Context, time.Time, time.Time) ([]domain.RepairAppointment, error) {
	return nil, nil
}

func (f *fakeRepo) SetWorkOrder(context.Context, uuid.UUID, uuid.UUID, time.Time) error { return nil }

type fakeRefs struct{}

func (fakeRefs) CustomerExists(context.Context, uuid.UUID) (bool, error)  { return true, nil }
func (fakeRefs) VehicleExists(context.Context, uuid.UUID) (bool, error)   { return true, nil }
func (fakeRefs) WarehouseExists(context.Context, uuid.UUID) (bool, error) { return true, nil }
func (fakeRefs) PartExists(context.Context, uuid.UUID) (bool, error)      { return true, nil }
func (fakeRefs) WorkExists(context.Context, uuid.UUID) (bool, error)      { return true, nil }

type fakePublisher struct {
	mu    sync.Mutex
	calls []*domain.RepairAppointment
	err   error
}

func (f *fakePublisher) Publish(_ context.Context, a *domain.RepairAppointment) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, a)
	return f.err
}

func (f *fakePublisher) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

func testCreateInput() domain.CreateInput {
	now := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	return domain.CreateInput{
		CustomerID:     uuid.New(),
		VehicleID:      uuid.New(),
		ScheduledStart: now,
		ScheduledEnd:   now.Add(time.Hour),
		Complaint:      "тест",
	}
}

func TestCreatePublishesEvent(t *testing.T) {
	repo := &fakeRepo{nextNum: "RA-TEST-1"}
	pub := &fakePublisher{}
	svc := NewRepairAppointmentService(repo, fakeRefs{}, nil, pub)

	a, err := svc.Create(context.Background(), testCreateInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if pub.count() != 1 {
		t.Fatalf("publisher calls = %d, want 1", pub.count())
	}
	published := pub.calls[0]
	if published.ID != a.ID {
		t.Errorf("published appointment id = %s, want %s", published.ID, a.ID)
	}
	if published.AppointmentNumber != "RA-TEST-1" {
		t.Errorf("published number = %q, want RA-TEST-1", published.AppointmentNumber)
	}
}

func TestCreateSkipsPublisherWhenNil(t *testing.T) {
	repo := &fakeRepo{nextNum: "RA-TEST-2"}
	svc := NewRepairAppointmentService(repo, fakeRefs{}, nil, nil)

	if _, err := svc.Create(context.Background(), testCreateInput()); err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestCreatePublishFailureDoesNotFailCreate(t *testing.T) {
	repo := &fakeRepo{nextNum: "RA-TEST-3"}
	pub := &fakePublisher{err: errors.New("kafka down")}
	svc := NewRepairAppointmentService(repo, fakeRefs{}, nil, pub)

	if _, err := svc.Create(context.Background(), testCreateInput()); err != nil {
		t.Fatalf("Create should succeed despite publish failure, got: %v", err)
	}
	if pub.count() != 1 {
		t.Fatalf("publisher calls = %d, want 1", pub.count())
	}
}
