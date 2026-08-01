//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestReferenceRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewReferenceRepository(testPool)

	now := time.Now().UTC()
	customerID := uuid.New()
	vehicleID := uuid.New()
	vin := "TESTVIN1234567890"
	if _, err := testPool.Exec(ctx,
		"INSERT INTO customers.customers (id, name, email, created_at, updated_at) VALUES ($1,$2,$3,$4,$5)",
		customerID, "Тест Клиент", "test@ref.local", now, now); err != nil {
		t.Fatalf("seed customer: %v", err)
	}
	if _, err := testPool.Exec(ctx,
		"INSERT INTO vehicles.vehicles (id, vin, make, model, year, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)",
		vehicleID, vin, "Toyota", "Camry", 2022, "available", now, now); err != nil {
		t.Fatalf("seed vehicle: %v", err)
	}

	exists, err := repo.CustomerExists(ctx, customerID)
	if err != nil {
		t.Fatalf("CustomerExists: %v", err)
	}
	if !exists {
		t.Fatal("customer should exist")
	}
	if exists, _ := repo.CustomerExists(ctx, uuid.New()); exists {
		t.Fatal("random customer should not exist")
	}

	exists, err = repo.VehicleExists(ctx, vehicleID)
	if err != nil {
		t.Fatalf("VehicleExists: %v", err)
	}
	if !exists {
		t.Fatal("vehicle should exist")
	}

	id, err := repo.LookupVehicleIDByVIN(ctx, "  "+vin+"  ")
	if err != nil {
		t.Fatalf("LookupVehicleIDByVIN: %v", err)
	}
	if id == nil || *id != vehicleID {
		t.Fatalf("LookupVehicleIDByVIN mismatch: %v", id)
	}
	if id, _ := repo.LookupVehicleIDByVIN(ctx, "NOPE"); id != nil {
		t.Fatalf("unknown vin should return nil, got %v", id)
	}
	if id, _ := repo.LookupVehicleIDByVIN(ctx, "   "); id != nil {
		t.Fatalf("blank vin should return nil, got %v", id)
	}

	name := repo.CustomerName(ctx, customerID)
	if name != "Тест Клиент" {
		t.Fatalf("CustomerName mismatch: %q", name)
	}

	gotVIN, label := repo.VehicleInfo(ctx, vehicleID)
	if gotVIN != vin || label != "Toyota Camry 2022" {
		t.Fatalf("VehicleInfo mismatch: %q %q", gotVIN, label)
	}
}
