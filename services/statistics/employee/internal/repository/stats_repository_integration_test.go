//go:build integration

package repository

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/dealer/dealer/employee-statistics-service/internal/domain"
)

func TestStatsRepository_InsertDealEvent(t *testing.T) {
	ctx := context.Background()
	repo := NewStatsRepository(testPool)

	dealID := uuid.New()
	customerID := uuid.New()
	vehicleID := uuid.New()
	occurredAt := time.Now().UTC()

	if err := repo.InsertDealEvent(ctx, dealID, customerID, vehicleID, "1250.00", "paid", occurredAt); err != nil {
		t.Fatalf("InsertDealEvent: %v", err)
	}
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM deal_events WHERE deal_id = $1`, dealID)
	})

	var count int64
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM deal_events WHERE deal_id = $1`, dealID).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row after insert, got %d", count)
	}

	if err := repo.InsertDealEvent(ctx, dealID, customerID, vehicleID, "9999.00", "completed", occurredAt); err != nil {
		t.Fatalf("re-insert: %v", err)
	}
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM deal_events WHERE deal_id = $1`, dealID).Scan(&count); err != nil {
		t.Fatalf("count after re-insert: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected re-insert to be a no-op, got %d rows", count)
	}
}

func TestStatsRepository_InsertDealEvent_InvalidStage(t *testing.T) {
	ctx := context.Background()
	repo := NewStatsRepository(testPool)

	err := repo.InsertDealEvent(ctx, uuid.New(), uuid.New(), uuid.New(), "1.00", "draft", time.Now().UTC())
	if err == nil {
		t.Fatal("expected error for stage 'draft', got nil")
	}
}

func TestStatsRepository_GetOverview(t *testing.T) {
	ctx := context.Background()
	repo := NewStatsRepository(testPool)

	var baseCount int64
	var baseSum float64
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*), COALESCE(SUM(amount), 0) FROM deal_events`).Scan(&baseCount, &baseSum); err != nil {
		t.Fatalf("baseline: %v", err)
	}

	seeds := []struct {
		id     uuid.UUID
		amount string
		stage  string
	}{
		{uuid.New(), "100.00", "paid"},
		{uuid.New(), "200.50", "paid"},
		{uuid.New(), "50.25", "completed"},
	}
	ids := make([]uuid.UUID, 0, len(seeds))
	for _, s := range seeds {
		if err := repo.InsertDealEvent(ctx, s.id, uuid.New(), uuid.New(), s.amount, s.stage, time.Now().UTC()); err != nil {
			t.Fatalf("seed %s: %v", s.stage, err)
		}
		ids = append(ids, s.id)
	}
	t.Cleanup(func() {
		for _, id := range ids {
			_, _ = testPool.Exec(ctx, `DELETE FROM deal_events WHERE deal_id = $1`, id)
		}
	})

	var ov *domain.Overview
	var err error
	ov, err = repo.GetOverview(ctx)
	if err != nil {
		t.Fatalf("GetOverview: %v", err)
	}
	if ov.DealsCount != baseCount+3 {
		t.Fatalf("DealsCount = %d, want %d", ov.DealsCount, baseCount+3)
	}
	if math.Abs(ov.TotalRevenue-(baseSum+350.75)) > 1e-6 {
		t.Fatalf("TotalRevenue = %v, want %v", ov.TotalRevenue, baseSum+350.75)
	}
	byStage := make(map[string]int64, len(ov.DealsByStage))
	for _, s := range ov.DealsByStage {
		byStage[s.Stage] = s.Count
	}
	if byStage["paid"] != 2 {
		t.Fatalf("DealsByStage paid = %d, want 2 (got %v)", byStage["paid"], ov.DealsByStage)
	}
	if byStage["completed"] != 1 {
		t.Fatalf("DealsByStage completed = %d, want 1 (got %v)", byStage["completed"], ov.DealsByStage)
	}
}
