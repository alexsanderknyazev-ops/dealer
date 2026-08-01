//go:build integration

package repository

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/appointments/internal/domain"
)

var apptSeq int64

func newTestAppointment() *domain.RepairAppointment {
	now := time.Now().UTC()
	base := time.Date(2055, 1, 1, 8, 0, 0, 0, time.UTC).Add(time.Duration(atomic.AddInt64(&apptSeq, 1)*2) * time.Hour)
	start := base
	end := start.Add(time.Hour)
	return &domain.RepairAppointment{
		ID:             uuid.New(),
		CustomerID:     uuid.New(),
		VehicleID:      uuid.New(),
		ScheduledStart: start,
		ScheduledEnd:   end,
		Status:         "scheduled",
		Complaint:      "замена масла",
		Notes:          "test appointment",
		CreatedAt:      now,
		UpdatedAt:      now,
		Labor: []domain.RepairAppointmentLabor{
			{
				ID:          uuid.New(),
				Description: "Замена масла",
				Quantity:    "1.0",
				UnitPrice:   "1500.00",
				SortOrder:   1,
				CreatedAt:   now,
			},
		},
		Parts: []domain.RepairAppointmentPart{
			{
				ID:          uuid.New(),
				PartID:      uuid.New(),
				WarehouseID: uuid.New(),
				Quantity:    5,
				UnitPrice:   "400.00",
				SortOrder:   1,
				CreatedAt:   now,
			},
		},
	}
}

func TestRepairAppointmentRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewRepairAppointmentRepository(testPool)

	a := newTestAppointment()
	num, err := repo.NextNumber(ctx)
	if err != nil {
		t.Fatalf("NextNumber: %v", err)
	}
	a.AppointmentNumber = num
	if err := repo.Create(ctx, a); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, a.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.AppointmentNumber != num || got.Status != "scheduled" || len(got.Labor) != 1 || len(got.Parts) != 1 {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
	if got.Labor[0].UnitPrice != "1500.00" || got.Parts[0].UnitPrice != "400.00" {
		t.Fatalf("lines mismatch: %+v %+v", got.Labor[0], got.Parts[0])
	}

	a.Complaint = "диагностика"
	a.Status = "in_progress"
	a.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, a, false); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ = repo.GetByID(ctx, a.ID)
	if got.Complaint != "диагностика" || got.Status != "in_progress" {
		t.Fatalf("update not persisted: %+v", got)
	}
}

func TestRepairAppointmentRepository_ReplaceLines(t *testing.T) {
	ctx := context.Background()
	repo := NewRepairAppointmentRepository(testPool)

	a := newTestAppointment()
	num, _ := repo.NextNumber(ctx)
	a.AppointmentNumber = num
	if err := repo.Create(ctx, a); err != nil {
		t.Fatalf("Create: %v", err)
	}

	now := time.Now().UTC()
	a.Labor = []domain.RepairAppointmentLabor{
		{
			ID:          uuid.New(),
			Description: "Сход-развал",
			Quantity:    "1.0",
			UnitPrice:   "2500.00",
			SortOrder:   1,
			CreatedAt:   now,
		},
	}
	a.Parts = nil
	a.UpdatedAt = now
	if err := repo.Update(ctx, a, true); err != nil {
		t.Fatalf("Update replaceLines: %v", err)
	}
	got, _ := repo.GetByID(ctx, a.ID)
	if len(got.Labor) != 1 || got.Labor[0].Description != "Сход-развал" {
		t.Fatalf("labor replace failed: %+v", got.Labor)
	}
	if len(got.Parts) != 0 {
		t.Fatalf("parts should be empty after replace, got %+v", got.Parts)
	}
}

func TestRepairAppointmentRepository_HasOverlap(t *testing.T) {
	ctx := context.Background()
	repo := NewRepairAppointmentRepository(testPool)

	a := newTestAppointment()
	num, _ := repo.NextNumber(ctx)
	a.AppointmentNumber = num
	if err := repo.Create(ctx, a); err != nil {
		t.Fatalf("Create: %v", err)
	}

	overlap, err := repo.HasOverlap(ctx, a.ScheduledStart.Add(15*time.Minute), a.ScheduledEnd.Add(-15*time.Minute), nil)
	if err != nil {
		t.Fatalf("HasOverlap: %v", err)
	}
	if !overlap {
		t.Fatal("expected overlap for intersecting window")
	}

	noOverlap, err := repo.HasOverlap(ctx, a.ScheduledEnd.Add(time.Hour), a.ScheduledEnd.Add(2*time.Hour), nil)
	if err != nil {
		t.Fatalf("HasOverlap: %v", err)
	}
	if noOverlap {
		t.Fatal("expected no overlap for disjoint window")
	}

	selfExcluded, err := repo.HasOverlap(ctx, a.ScheduledStart, a.ScheduledEnd, &a.ID)
	if err != nil {
		t.Fatalf("HasOverlap excludeID: %v", err)
	}
	if selfExcluded {
		t.Fatal("self should be excluded")
	}

	busy, err := repo.ListBusyInRange(ctx, a.ScheduledStart.Add(-time.Hour), a.ScheduledEnd.Add(time.Hour))
	if err != nil {
		t.Fatalf("ListBusyInRange: %v", err)
	}
	found := false
	for _, it := range busy {
		if it.ID == a.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("appointment should be busy in range")
	}

	cancelled := newTestAppointment()
	cancelled.ScheduledStart = a.ScheduledEnd.Add(2 * time.Hour)
	cancelled.ScheduledEnd = cancelled.ScheduledStart.Add(30 * time.Minute)
	cancelled.Status = "cancelled"
	cNum, _ := repo.NextNumber(ctx)
	cancelled.AppointmentNumber = cNum
	if err := repo.Create(ctx, cancelled); err != nil {
		t.Fatalf("Create cancelled: %v", err)
	}
	ov, err := repo.HasOverlap(ctx, cancelled.ScheduledStart, cancelled.ScheduledEnd, nil)
	if err != nil {
		t.Fatalf("HasOverlap cancelled: %v", err)
	}
	if ov {
		t.Fatal("cancelled appointment should not overlap")
	}
}

func TestRepairAppointmentRepository_SetWorkOrder(t *testing.T) {
	ctx := context.Background()
	repo := NewRepairAppointmentRepository(testPool)

	a := newTestAppointment()
	num, _ := repo.NextNumber(ctx)
	a.AppointmentNumber = num
	if err := repo.Create(ctx, a); err != nil {
		t.Fatalf("Create: %v", err)
	}

	woID := uuid.New()
	if err := repo.SetWorkOrder(ctx, a.ID, woID, time.Now().UTC()); err != nil {
		t.Fatalf("SetWorkOrder: %v", err)
	}
	got, _ := repo.GetByID(ctx, a.ID)
	if got.WorkOrderID == nil || *got.WorkOrderID != woID || got.Status != "in_progress" {
		t.Fatalf("SetWorkOrder not persisted: %+v", got)
	}

	woID2 := uuid.New()
	if err := repo.SetWorkOrder(ctx, a.ID, woID2, time.Now().UTC()); err != nil {
		t.Fatalf("SetWorkOrder second: %v", err)
	}
	got, _ = repo.GetByID(ctx, a.ID)
	if got.WorkOrderID == nil || *got.WorkOrderID != woID {
		t.Fatalf("second SetWorkOrder should be a no-op: %+v", got)
	}
}

func TestRepairAppointmentRepository_List(t *testing.T) {
	ctx := context.Background()
	repo := NewRepairAppointmentRepository(testPool)

	from := time.Now().UTC().Add(20 * time.Hour)
	a1 := newTestAppointment()
	a1.ScheduledStart = from
	a1.ScheduledEnd = from.Add(time.Hour)
	a1.Status = "draft"
	n1, _ := repo.NextNumber(ctx)
	a1.AppointmentNumber = n1
	a2 := newTestAppointment()
	a2.ScheduledStart = from.Add(2 * time.Hour)
	a2.ScheduledEnd = from.Add(3 * time.Hour)
	n2, _ := repo.NextNumber(ctx)
	a2.AppointmentNumber = n2
	for _, a := range []*domain.RepairAppointment{a1, a2} {
		if err := repo.Create(ctx, a); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	list, total, err := repo.List(ctx, 10, 0, "draft", nil, nil)
	if err != nil {
		t.Fatalf("List by status: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("List by status draft: total=%d len=%d", total, len(list))
	}

	fromP := from.Add(-time.Hour)
	toP := from.Add(90 * time.Minute)
	list, total, err = repo.List(ctx, 10, 0, "", &fromP, &toP)
	if err != nil {
		t.Fatalf("List by range: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("List by range: total=%d len=%d", total, len(list))
	}

	list, _, err = repo.List(ctx, 1, 0, "", nil, nil)
	if err != nil {
		t.Fatalf("List paginated: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("pagination limit not honored, len=%d", len(list))
	}
}

func TestRepairAppointmentRepository_InvalidStatus(t *testing.T) {
	ctx := context.Background()
	repo := NewRepairAppointmentRepository(testPool)

	a := newTestAppointment()
	n, _ := repo.NextNumber(ctx)
	a.AppointmentNumber = n
	a.Status = "nope"
	if err := repo.Create(ctx, a); err == nil {
		t.Fatal("expected CHECK violation for invalid status")
	}
}

func TestRepairAppointmentRepository_MissingRow(t *testing.T) {
	ctx := context.Background()
	repo := NewRepairAppointmentRepository(testPool)
	if _, err := repo.GetByID(ctx, uuid.New()); err != pgx.ErrNoRows {
		t.Fatalf("expected ErrNoRows, got %v", err)
	}
}
