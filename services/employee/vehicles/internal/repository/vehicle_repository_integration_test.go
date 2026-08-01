//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/dealer/dealer/services/vehicles/internal/domain"
)

func TestVehicleRepository_CRUD(t *testing.T) {
	repo := NewVehicleRepository(testPool)
	ctx := context.Background()

	v := &domain.Vehicle{
		ID:        uuid.New(),
		VIN:       "ITVIN000000000001",
		Make:      "ITMake",
		Model:     "ITModel",
		Year:      2022,
		MileageKm: 15000,
		Price:     "2500000.00",
		Status:    "available",
		Color:     "black",
		Notes:     "integration test",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := repo.Create(ctx, v); err != nil {
		t.Fatalf("Create: %v", err)
	}

	byID, err := repo.GetByID(ctx, v.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if byID.VIN != v.VIN || byID.Make != v.Make || byID.Price != v.Price || byID.Status != v.Status {
		t.Fatalf("GetByID mismatch: %+v", byID)
	}

	byVIN, err := repo.GetByVIN(ctx, v.VIN)
	if err != nil {
		t.Fatalf("GetByVIN: %v", err)
	}
	if byVIN.ID != v.ID {
		t.Fatalf("GetByVIN mismatch: %+v", byVIN)
	}

	if _, err := repo.GetByID(ctx, uuid.New()); err != pgx.ErrNoRows {
		t.Fatalf("GetByID(missing): got %v want pgx.ErrNoRows", err)
	}
	if _, err := repo.GetByVIN(ctx, "MISSINGVIN"); err != pgx.ErrNoRows {
		t.Fatalf("GetByVIN(missing): got %v want pgx.ErrNoRows", err)
	}

	v.Status = "sold"
	v.MileageKm = 16000
	v.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, v); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err := repo.GetByID(ctx, v.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Status != "sold" || got.MileageKm != 16000 {
		t.Fatalf("Update not persisted: %+v", got)
	}

	if err := repo.Delete(ctx, v.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, v.ID); err != pgx.ErrNoRows {
		t.Fatalf("GetByID after delete: got %v want pgx.ErrNoRows", err)
	}
}

func TestVehicleRepository_ListFilters(t *testing.T) {
	repo := NewVehicleRepository(testPool)
	ctx := context.Background()

	brandA := uuid.New()
	dealerPoint := uuid.New()
	warehouse := uuid.New()

	vehicles := []*domain.Vehicle{
		{ID: uuid.New(), VIN: "ITFILTER00000001", Make: "ITFilterMake", Model: "ModelA", Year: 2020, MileageKm: 10000, Price: "1000000.00", Status: "available", BrandID: &brandA, DealerPointID: &dealerPoint, WarehouseID: &warehouse, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		{ID: uuid.New(), VIN: "ITFILTER00000002", Make: "ITFilterMake", Model: "ModelB", Year: 2021, MileageKm: 20000, Price: "2000000.00", Status: "sold", BrandID: &brandA, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		{ID: uuid.New(), VIN: "ITFILTER00000003", Make: "OtherMake", Model: "ModelC", Year: 2022, MileageKm: 30000, Price: "3000000.00", Status: "reserved", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
	}
	for _, v := range vehicles {
		if err := repo.Create(ctx, v); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	list, total, err := repo.List(ctx, domain.VehicleListFilter{Limit: 10, Offset: 0, Search: "ITFilterMake"})
	if err != nil {
		t.Fatalf("List search: %v", err)
	}
	if total < 2 || len(list) < 2 {
		t.Fatalf("List search: total=%d len=%d want >=2", total, len(list))
	}

	list, total, err = repo.List(ctx, domain.VehicleListFilter{Limit: 10, Offset: 0, StatusFilter: "sold"})
	if err != nil {
		t.Fatalf("List status: %v", err)
	}
	if total < 1 {
		t.Fatalf("List status total: got %d want >=1", total)
	}
	for _, v := range list {
		if v.Status != "sold" {
			t.Fatalf("List status filter returned %q", v.Status)
		}
	}

	list, total, err = repo.List(ctx, domain.VehicleListFilter{Limit: 10, Offset: 0, BrandID: &brandA})
	if err != nil {
		t.Fatalf("List brand: %v", err)
	}
	if total < 2 || len(list) < 2 {
		t.Fatalf("List brand: total=%d len=%d want >=2", total, len(list))
	}
	for _, v := range list {
		if v.BrandID == nil || *v.BrandID != brandA {
			t.Fatalf("List brand filter returned %+v", v)
		}
	}

	list, total, err = repo.List(ctx, domain.VehicleListFilter{Limit: 10, Offset: 0, DealerPointID: &dealerPoint})
	if err != nil {
		t.Fatalf("List dealer point: %v", err)
	}
	if total < 1 {
		t.Fatalf("List dealer point total: got %d want >=1", total)
	}

	page, total, err := repo.List(ctx, domain.VehicleListFilter{Limit: 2, Offset: 0, Search: "ITFilterMake"})
	if err != nil {
		t.Fatalf("List paged: %v", err)
	}
	if len(page) > 2 {
		t.Fatalf("List page len: got %d want <=2", len(page))
	}
	if total < 2 {
		t.Fatalf("List page total: got %d want >=2", total)
	}

	page, total, err = repo.List(ctx, domain.VehicleListFilter{Limit: 2, Offset: 2, Search: "ITFilterMake"})
	if err != nil {
		t.Fatalf("List paged offset: %v", err)
	}
	if len(page) != 0 {
		t.Fatalf("List offset len: got %d want 0", len(page))
	}
	if total < 2 {
		t.Fatalf("List offset total: got %d want >=2", total)
	}
}

func TestVehicleRepository_UniqueVIN(t *testing.T) {
	repo := NewVehicleRepository(testPool)
	ctx := context.Background()

	v := &domain.Vehicle{
		ID: uuid.New(), VIN: "ITUNIQUEVIN0001", Make: "Unique", Model: "Vin", Year: 2020, MileageKm: 0, Price: "1.00", Status: "available", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, v); err != nil {
		t.Fatalf("Create: %v", err)
	}

	dup := *v
	dup.ID = uuid.New()
	err := repo.Create(ctx, &dup)
	if err == nil {
		t.Fatal("expected unique violation for duplicate VIN")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("expected SQLSTATE 23505, got %v", err)
	}
}
