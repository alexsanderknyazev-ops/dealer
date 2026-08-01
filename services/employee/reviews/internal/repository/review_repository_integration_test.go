//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/employee-reviews-service/internal/domain"
)

func newEmployeeReview(clientID uuid.UUID, rating int32, status string, occurred time.Time) *domain.Review {
	return &domain.Review{
		ID: uuid.New(), ReviewID: uuid.New(), ClientID: clientID, UserID: uuid.New(),
		ClientEmail: "emp.review@example.com", ClientFullName: "Emp Review",
		DealerPointID: uuid.New(), VehicleID: uuid.New(), VehicleVIN: "ITVINEMP0000000001",
		VehicleMake: "Toyota", VehicleModel: "Camry", VehicleYear: 2021,
		Rating: rating, Text: "text", Status: status, OccurredAt: occurred, CreatedAt: time.Now().UTC(),
	}
}

func TestEmployeeReviewRepository_InsertIdempotent(t *testing.T) {
	repo := NewReviewRepository(testPool)
	ctx := context.Background()

	rv := newEmployeeReview(uuid.New(), 5, "published", time.Now().UTC())
	if err := repo.Insert(ctx, rv); err != nil {
		t.Fatalf("first Insert: %v", err)
	}

	dup := newEmployeeReview(rv.ClientID, 1, "rejected", time.Now().UTC())
	dup.ReviewID = rv.ReviewID
	if err := repo.Insert(ctx, dup); err != nil {
		t.Fatalf("second Insert same review_id: %v", err)
	}

	var count int
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM reviews WHERE review_id = $1`, rv.ReviewID).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("rows for review_id: got %d want 1", count)
	}

	var rowID uuid.UUID
	if err := testPool.QueryRow(ctx, `SELECT id FROM reviews WHERE review_id = $1`, rv.ReviewID).Scan(&rowID); err != nil {
		t.Fatalf("fetch row id: %v", err)
	}
	got, err := repo.GetByID(ctx, rowID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Rating != 5 || got.Status != "published" {
		t.Fatalf("original row must be kept: %+v", got)
	}

	if _, err := repo.GetByID(ctx, uuid.New()); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("GetByID missing: got %v want pgx.ErrNoRows", err)
	}
}

func TestEmployeeReviewRepository_List(t *testing.T) {
	repo := NewReviewRepository(testPool)
	ctx := context.Background()

	clientID := uuid.New()
	dealerA := uuid.New()
	dealerB := uuid.New()
	now := time.Now().UTC()

	reviews := []*domain.Review{
		newEmployeeReview(clientID, 5, "published", now.Add(-3*time.Hour)),
		newEmployeeReview(clientID, 4, "draft", now.Add(-2*time.Hour)),
		newEmployeeReview(clientID, 3, "published", now.Add(-time.Hour)),
	}
	reviews[0].DealerPointID = dealerA
	reviews[1].DealerPointID = dealerA
	reviews[2].DealerPointID = dealerB
	for _, rv := range reviews {
		if err := repo.Insert(ctx, rv); err != nil {
			t.Fatalf("Insert: %v", err)
		}
	}

	list, total, err := repo.List(ctx, domain.ReviewListParams{ClientID: &clientID})
	if err != nil {
		t.Fatalf("List by client: %v", err)
	}
	if total != 3 || len(list) != 3 {
		t.Fatalf("List by client: total=%d len=%d want 3/3", total, len(list))
	}
	if list[0].Rating != 3 {
		t.Fatalf("List ordering: first rating got %d want 3 (occurred_at DESC)", list[0].Rating)
	}

	list, total, err = repo.List(ctx, domain.ReviewListParams{DealerPointID: &dealerA})
	if err != nil {
		t.Fatalf("List by dealer: %v", err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("List by dealer: total=%d len=%d want 2/2", total, len(list))
	}

	list, total, err = repo.List(ctx, domain.ReviewListParams{ClientID: &clientID, Status: "published"})
	if err != nil {
		t.Fatalf("List by client+status: %v", err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("List by client+status: total=%d len=%d want 2/2", total, len(list))
	}

	list, total, err = repo.List(ctx, domain.ReviewListParams{Limit: 1, Offset: 0})
	if err != nil {
		t.Fatalf("List paged: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("List limit=1: got len %d want 1", len(list))
	}
	if total < 3 {
		t.Fatalf("List total: got %d want >=3", total)
	}
}

func TestEmployeeReviewRepository_LimitNormalization(t *testing.T) {
	repo := NewReviewRepository(testPool)
	ctx := context.Background()

	clientID := uuid.New()
	occurred := time.Now().UTC()
	for i := 0; i < 205; i++ {
		rv := newEmployeeReview(clientID, 5, "published", occurred.Add(-time.Duration(i)*time.Second))
		if err := repo.Insert(ctx, rv); err != nil {
			t.Fatalf("Insert %d: %v", i, err)
		}
	}

	list, total, err := repo.List(ctx, domain.ReviewListParams{ClientID: &clientID, Limit: 1000, Offset: 0})
	if err != nil {
		t.Fatalf("List limit 1000: %v", err)
	}
	if total != 205 {
		t.Fatalf("total: got %d want 205", total)
	}
	if len(list) != 200 {
		t.Fatalf("clamped limit: got %d want 200", len(list))
	}

	list, total, err = repo.List(ctx, domain.ReviewListParams{ClientID: &clientID, Limit: 0})
	if err != nil {
		t.Fatalf("List default limit: %v", err)
	}
	if total != 205 {
		t.Fatalf("default total: got %d want 205", total)
	}
	if len(list) != 50 {
		t.Fatalf("default limit: got %d want 50", len(list))
	}
}

func TestEmployeeReviewRepository_Stats(t *testing.T) {
	repo := NewReviewRepository(testPool)
	ctx := context.Background()

	clientID := uuid.New()
	now := time.Now().UTC()
	reviews := []*domain.Review{
		newEmployeeReview(clientID, 4, "published", now.Add(-2*time.Hour)),
		newEmployeeReview(clientID, 5, "published", now.Add(-time.Hour)),
		newEmployeeReview(clientID, 1, "draft", now),
	}
	for _, rv := range reviews {
		if err := repo.Insert(ctx, rv); err != nil {
			t.Fatalf("Insert: %v", err)
		}
	}

	stats, err := repo.Stats(ctx, &clientID, nil)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.TotalCount != 3 {
		t.Fatalf("TotalCount: got %d want 3", stats.TotalCount)
	}
	if stats.AverageRating != 4.5 {
		t.Fatalf("AverageRating: got %v want 4.5", stats.AverageRating)
	}

	all, err := repo.Stats(ctx, nil, nil)
	if err != nil {
		t.Fatalf("Stats all: %v", err)
	}
	if all.TotalCount < 3 {
		t.Fatalf("Stats all TotalCount: got %d want >=3", all.TotalCount)
	}
	if len(all.ByStatus) == 0 {
		t.Fatal("Stats all ByStatus: want non-empty")
	}
}
