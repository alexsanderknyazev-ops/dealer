//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/brands/internal/domain"
)

func TestLaborRateRepository_UpsertGetAndDelete(t *testing.T) {
	brandRepo := NewBrandRepository(testPool)
	ctx := context.Background()

	brand := &domain.Brand{
		ID:        uuid.New(),
		Name:      "IT Rate Brand",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := brandRepo.Create(ctx, brand); err != nil {
		t.Fatalf("Create brand: %v", err)
	}

	repo := NewLaborRateRepository(testPool)
	dealerPointID := uuid.New()
	rate := &domain.BrandLaborRate{
		ID:                  uuid.New(),
		BrandID:             brand.ID,
		DealerPointID:       dealerPointID,
		WarrantyHourPrice:   "1500.00",
		CommercialHourPrice: "1800.00",
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}
	if err := repo.Upsert(ctx, rate); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	byID, err := repo.GetByID(ctx, rate.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if byID.WarrantyHourPrice != "1500.00" || byID.CommercialHourPrice != "1800.00" {
		t.Fatalf("GetByID mismatch: %+v", byID)
	}

	byPair, err := repo.GetByBrandAndDealerPoint(ctx, brand.ID, dealerPointID)
	if err != nil {
		t.Fatalf("GetByBrandAndDealerPoint: %v", err)
	}
	if byPair.ID != rate.ID {
		t.Fatalf("GetByBrandAndDealerPoint mismatch: %+v", byPair)
	}

	if _, err := repo.GetByID(ctx, uuid.New()); err != pgx.ErrNoRows {
		t.Fatalf("GetByID(missing): got %v want pgx.ErrNoRows", err)
	}
	if _, err := repo.GetByBrandAndDealerPoint(ctx, uuid.New(), uuid.New()); err != pgx.ErrNoRows {
		t.Fatalf("GetByBrandAndDealerPoint(missing): got %v want pgx.ErrNoRows", err)
	}

	rate.WarrantyHourPrice = "1600.00"
	rate.UpdatedAt = time.Now().UTC()
	if err := repo.Upsert(ctx, rate); err != nil {
		t.Fatalf("Upsert update: %v", err)
	}
	byPair, err = repo.GetByBrandAndDealerPoint(ctx, brand.ID, dealerPointID)
	if err != nil {
		t.Fatalf("GetByBrandAndDealerPoint after update: %v", err)
	}
	if byPair.WarrantyHourPrice != "1600.00" {
		t.Fatalf("Upsert not applied: %+v", byPair)
	}

	if err := repo.Delete(ctx, rate.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, rate.ID); err != pgx.ErrNoRows {
		t.Fatalf("GetByID after delete: got %v want pgx.ErrNoRows", err)
	}
}

func TestLaborRateRepository_ListFilters(t *testing.T) {
	brandRepo := NewBrandRepository(testPool)
	ctx := context.Background()

	brandA := &domain.Brand{ID: uuid.New(), Name: "IT Rate Brand A", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	brandB := &domain.Brand{ID: uuid.New(), Name: "IT Rate Brand B", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	for _, b := range []*domain.Brand{brandA, brandB} {
		if err := brandRepo.Create(ctx, b); err != nil {
			t.Fatalf("Create brand: %v", err)
		}
	}

	repo := NewLaborRateRepository(testPool)
	dealerPointID := uuid.New()
	rates := []*domain.BrandLaborRate{
		{ID: uuid.New(), BrandID: brandA.ID, DealerPointID: dealerPointID, WarrantyHourPrice: "1000.00", CommercialHourPrice: "1200.00", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		{ID: uuid.New(), BrandID: brandB.ID, DealerPointID: dealerPointID, WarrantyHourPrice: "2000.00", CommercialHourPrice: "2400.00", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
	}
	for _, rate := range rates {
		if err := repo.Upsert(ctx, rate); err != nil {
			t.Fatalf("Upsert: %v", err)
		}
	}

	list, total, err := repo.List(ctx, 10, 0, &brandA.ID, nil)
	if err != nil {
		t.Fatalf("List brand: %v", err)
	}
	if total < 1 || len(list) < 1 {
		t.Fatalf("List brand: total=%d len=%d want >=1", total, len(list))
	}
	for _, r := range list {
		if r.BrandID != brandA.ID {
			t.Fatalf("List brand filter returned %+v", r)
		}
	}

	list, total, err = repo.List(ctx, 10, 0, nil, &dealerPointID)
	if err != nil {
		t.Fatalf("List dealer point: %v", err)
	}
	if total < 2 {
		t.Fatalf("List dealer point total: got %d want >=2", total)
	}

	page, total, err := repo.List(ctx, 1, 0, nil, nil)
	if err != nil {
		t.Fatalf("List paged: %v", err)
	}
	if len(page) > 1 {
		t.Fatalf("List page len: got %d want <=1", len(page))
	}
	if total < 2 {
		t.Fatalf("List page total: got %d want >=2", total)
	}

	missingBrand := uuid.New()
	list, total, err = repo.List(ctx, 10, 0, &missingBrand, nil)
	if err != nil {
		t.Fatalf("List missing brand: %v", err)
	}
	if total != 0 || len(list) != 0 {
		t.Fatalf("List missing brand: total=%d len=%d want 0", total, len(list))
	}
}
