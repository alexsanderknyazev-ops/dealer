//go:build integration

package repository

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/dealer/dealer/client-statistics-service/internal/domain"
)

func TestStatsRepository_InsertClientRegistration(t *testing.T) {
	ctx := context.Background()
	repo := NewStatsRepository(testPool)

	userID := uuid.New()
	vehicleID := uuid.New()
	occurredAt := time.Now().UTC()

	if err := repo.InsertClientRegistration(ctx, userID, "test@example.com", &vehicleID, occurredAt); err != nil {
		t.Fatalf("InsertClientRegistration: %v", err)
	}
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM client_registration_events WHERE user_id = $1`, userID)
	})

	var count int64
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM client_registration_events WHERE user_id = $1`, userID).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row after insert, got %d", count)
	}

	if err := repo.InsertClientRegistration(ctx, userID, "other@example.com", &vehicleID, occurredAt); err != nil {
		t.Fatalf("re-insert: %v", err)
	}
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM client_registration_events WHERE user_id = $1`, userID).Scan(&count); err != nil {
		t.Fatalf("count after re-insert: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected re-insert to be a no-op, got %d rows", count)
	}
}

func TestStatsRepository_InsertClientRegistration_NilVehicle(t *testing.T) {
	ctx := context.Background()
	repo := NewStatsRepository(testPool)

	userID := uuid.New()
	if err := repo.InsertClientRegistration(ctx, userID, "novi@example.com", nil, time.Now().UTC()); err != nil {
		t.Fatalf("InsertClientRegistration with nil vehicle: %v", err)
	}
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM client_registration_events WHERE user_id = $1`, userID)
	})

	var count int64
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM client_registration_events WHERE user_id = $1`, userID).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}
}

func TestStatsRepository_InsertReviewEvent(t *testing.T) {
	ctx := context.Background()
	repo := NewStatsRepository(testPool)

	reviewID := uuid.New()
	clientID := uuid.New()
	userID := uuid.New()
	dealerPointID := uuid.New()
	vehicleID := uuid.New()
	occurredAt := time.Now().UTC()

	if err := repo.InsertReviewEvent(ctx, reviewID, clientID, userID, dealerPointID, vehicleID, 5, "published", occurredAt); err != nil {
		t.Fatalf("InsertReviewEvent: %v", err)
	}
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM review_events WHERE review_id = $1`, reviewID)
	})

	var count int64
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM review_events WHERE review_id = $1`, reviewID).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row after insert, got %d", count)
	}

	if err := repo.InsertReviewEvent(ctx, reviewID, clientID, userID, dealerPointID, vehicleID, 3, "draft", occurredAt); err != nil {
		t.Fatalf("re-insert: %v", err)
	}
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM review_events WHERE review_id = $1`, reviewID).Scan(&count); err != nil {
		t.Fatalf("count after re-insert: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected re-insert to be a no-op, got %d rows", count)
	}
}

func TestStatsRepository_InsertReviewEvent_InvalidRating(t *testing.T) {
	ctx := context.Background()
	repo := NewStatsRepository(testPool)

	err := repo.InsertReviewEvent(ctx, uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), 0, "published", time.Now().UTC())
	if err == nil {
		t.Fatal("expected error for rating 0, got nil")
	}
	err = repo.InsertReviewEvent(ctx, uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), 6, "published", time.Now().UTC())
	if err == nil {
		t.Fatal("expected error for rating 6, got nil")
	}
}

func TestStatsRepository_GetOverview(t *testing.T) {
	ctx := context.Background()
	repo := NewStatsRepository(testPool)

	var baseClients, baseReviews, baseVehicles int64
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM client_registration_events`).Scan(&baseClients); err != nil {
		t.Fatalf("baseline clients: %v", err)
	}
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM review_events`).Scan(&baseReviews); err != nil {
		t.Fatalf("baseline reviews: %v", err)
	}
	if err := testPool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT vehicle_id) FROM (
			SELECT vehicle_id FROM client_registration_events WHERE vehicle_id IS NOT NULL
			UNION
			SELECT vehicle_id FROM review_events
		) vehicles
	`).Scan(&baseVehicles); err != nil {
		t.Fatalf("baseline vehicles: %v", err)
	}
	var basePubCount int64
	var basePubAvg float64
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*), COALESCE(AVG(rating), 0) FROM review_events WHERE status = 'published'`).Scan(&basePubCount, &basePubAvg); err != nil {
		t.Fatalf("baseline published: %v", err)
	}
	baseStatus := make(map[string]int64)
	rows, err := testPool.Query(ctx, `SELECT status, COUNT(*) FROM review_events GROUP BY status`)
	if err != nil {
		t.Fatalf("baseline statuses: %v", err)
	}
	for rows.Next() {
		var s string
		var c int64
		if err := rows.Scan(&s, &c); err != nil {
			t.Fatalf("baseline status scan: %v", err)
		}
		baseStatus[s] = c
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatalf("baseline statuses: %v", err)
	}

	u1, u2, u3 := uuid.New(), uuid.New(), uuid.New()
	v1, v2, v3 := uuid.New(), uuid.New(), uuid.New()
	dp := uuid.New()
	client1, client2 := uuid.New(), uuid.New()
	now := time.Now().UTC()

	regIDs := []uuid.UUID{u1, u2, u3}
	reviewIDs := make([]uuid.UUID, 0, 2)

	v1p, v2p := v1, v2
	if err := repo.InsertClientRegistration(ctx, u1, "one@test.local", &v1p, now); err != nil {
		t.Fatalf("seed registration 1: %v", err)
	}
	if err := repo.InsertClientRegistration(ctx, u2, "two@test.local", &v2p, now); err != nil {
		t.Fatalf("seed registration 2: %v", err)
	}
	if err := repo.InsertClientRegistration(ctx, u3, "three@test.local", nil, now); err != nil {
		t.Fatalf("seed registration 3: %v", err)
	}
	if err := repo.InsertClientRegistration(ctx, u1, "one@test.local", &v1p, now); err != nil {
		t.Fatalf("seed duplicate registration: %v", err)
	}

	r1 := uuid.New()
	reviewIDs = append(reviewIDs, r1)
	if err := repo.InsertReviewEvent(ctx, r1, client1, u1, dp, v1, 5, "published", now); err != nil {
		t.Fatalf("seed review published: %v", err)
	}
	r2 := uuid.New()
	reviewIDs = append(reviewIDs, r2)
	if err := repo.InsertReviewEvent(ctx, r2, client2, u2, dp, v3, 2, "draft", now); err != nil {
		t.Fatalf("seed review draft: %v", err)
	}

	t.Cleanup(func() {
		for _, id := range regIDs {
			_, _ = testPool.Exec(ctx, `DELETE FROM client_registration_events WHERE user_id = $1`, id)
		}
		for _, id := range reviewIDs {
			_, _ = testPool.Exec(ctx, `DELETE FROM review_events WHERE review_id = $1`, id)
		}
	})

	var ov *domain.Overview
	ov, err = repo.GetOverview(ctx)
	if err != nil {
		t.Fatalf("GetOverview: %v", err)
	}

	if ov.ClientsCount != baseClients+3 {
		t.Fatalf("ClientsCount = %d, want %d", ov.ClientsCount, baseClients+3)
	}
	if ov.RegisteredUsersCount != baseClients+3 {
		t.Fatalf("RegisteredUsersCount = %d, want %d", ov.RegisteredUsersCount, baseClients+3)
	}
	if ov.ClientVehiclesCount != baseVehicles+3 {
		t.Fatalf("ClientVehiclesCount = %d, want %d", ov.ClientVehiclesCount, baseVehicles+3)
	}
	if ov.ReviewsCount != baseReviews+2 {
		t.Fatalf("ReviewsCount = %d, want %d", ov.ReviewsCount, baseReviews+2)
	}
	wantAvg := (basePubAvg*float64(basePubCount) + 5) / float64(basePubCount+1)
	if math.Abs(ov.AverageRating-wantAvg) > 1e-9 {
		t.Fatalf("AverageRating = %v, want %v", ov.AverageRating, wantAvg)
	}
	byStatus := make(map[string]int64, len(ov.ReviewsByStatus))
	for _, s := range ov.ReviewsByStatus {
		byStatus[s.Status] = s.Count
	}
	if byStatus["published"] != baseStatus["published"]+1 {
		t.Fatalf("published count = %d, want %d", byStatus["published"], baseStatus["published"]+1)
	}
	if byStatus["draft"] != baseStatus["draft"]+1 {
		t.Fatalf("draft count = %d, want %d", byStatus["draft"], baseStatus["draft"]+1)
	}
}
