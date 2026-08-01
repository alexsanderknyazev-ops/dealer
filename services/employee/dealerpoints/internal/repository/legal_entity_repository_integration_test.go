//go:build integration

package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/dealer/dealer/services/dealerpoints/internal/domain"
)

func TestLegalEntityRepository_CRUD(t *testing.T) {
	repo := NewLegalEntityRepository(testPool)
	ctx := context.Background()

	e := &domain.LegalEntity{
		ID:        uuid.New(),
		Name:      "IT Legal Entity",
		INN:       "7701000001",
		Address:   "Moscow",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, e); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, e.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != e.Name || got.INN != e.INN || got.Address != e.Address {
		t.Fatalf("GetByID mismatch: %+v", got)
	}

	if _, err := repo.GetByID(ctx, uuid.New()); err != pgx.ErrNoRows {
		t.Fatalf("GetByID(missing): got %v want pgx.ErrNoRows", err)
	}

	e.Name = "IT Legal Entity Renamed"
	e.Address = "Saint Petersburg"
	e.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, e); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err = repo.GetByID(ctx, e.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Name != "IT Legal Entity Renamed" || got.Address != "Saint Petersburg" {
		t.Fatalf("Update not persisted: %+v", got)
	}

	if err := repo.Delete(ctx, e.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, e.ID); err != pgx.ErrNoRows {
		t.Fatalf("GetByID after delete: got %v want pgx.ErrNoRows", err)
	}
}

func TestLegalEntityRepository_ListSearchAndPagination(t *testing.T) {
	repo := NewLegalEntityRepository(testPool)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		e := &domain.LegalEntity{
			ID:        uuid.New(),
			Name:      fmt.Sprintf("IT Legal Search %d", i),
			INN:       fmt.Sprintf("77020000%02d", i),
			Address:   "Moscow",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		if err := repo.Create(ctx, e); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	list, total, err := repo.List(ctx, 10, 0, "IT Legal Search")
	if err != nil {
		t.Fatalf("List search: %v", err)
	}
	if total < 3 || len(list) < 3 {
		t.Fatalf("List search: total=%d len=%d want >=3", total, len(list))
	}

	page, total, err := repo.List(ctx, 2, 0, "")
	if err != nil {
		t.Fatalf("List paged: %v", err)
	}
	if len(page) > 2 {
		t.Fatalf("List page len: got %d want <=2", len(page))
	}
	if total < 3 {
		t.Fatalf("List page total: got %d want >=3", total)
	}
}

func TestLegalEntityRepository_DealerPointLinks(t *testing.T) {
	dpRepo := NewDealerPointRepository(testPool)
	leRepo := NewLegalEntityRepository(testPool)
	ctx := context.Background()

	dp := &domain.DealerPoint{
		ID:        uuid.New(),
		Name:      "IT DP Links",
		Address:   "Moscow",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := dpRepo.Create(ctx, dp); err != nil {
		t.Fatalf("Create dealer point: %v", err)
	}

	le := &domain.LegalEntity{
		ID:        uuid.New(),
		Name:      "IT Linked Legal",
		INN:       "7703000001",
		Address:   "Moscow",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := leRepo.Create(ctx, le); err != nil {
		t.Fatalf("Create legal entity: %v", err)
	}

	if err := leRepo.LinkToDealerPoint(ctx, dp.ID, le.ID); err != nil {
		t.Fatalf("LinkToDealerPoint: %v", err)
	}
	list, total, err := leRepo.ListByDealerPoint(ctx, dp.ID, 10, 0)
	if err != nil {
		t.Fatalf("ListByDealerPoint: %v", err)
	}
	if total < 1 {
		t.Fatalf("ListByDealerPoint total: got %d want >=1", total)
	}
	found := false
	for _, e := range list {
		if e.ID == le.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("ListByDealerPoint missing linked entity: %+v", list)
	}

	if err := leRepo.UnlinkFromDealerPoint(ctx, dp.ID, le.ID); err != nil {
		t.Fatalf("UnlinkFromDealerPoint: %v", err)
	}
	list, total, err = leRepo.ListByDealerPoint(ctx, dp.ID, 10, 0)
	if err != nil {
		t.Fatalf("ListByDealerPoint after unlink: %v", err)
	}
	if total != 0 {
		t.Fatalf("ListByDealerPoint after unlink total: got %d want 0", total)
	}
}

func TestLegalEntityRepository_UniqueINN(t *testing.T) {
	repo := NewLegalEntityRepository(testPool)
	ctx := context.Background()

	e := &domain.LegalEntity{
		ID:        uuid.New(),
		Name:      "IT Unique INN",
		INN:       "7704000001",
		Address:   "Moscow",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, e); err != nil {
		t.Fatalf("Create: %v", err)
	}

	dup := *e
	dup.ID = uuid.New()
	err := repo.Create(ctx, &dup)
	if err == nil {
		t.Fatal("expected unique violation for duplicate INN")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("expected SQLSTATE 23505, got %v", err)
	}
}
