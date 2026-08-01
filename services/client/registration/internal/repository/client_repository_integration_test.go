//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/client-registration-service/internal/domain"
)

func TestClientRepository_CreateClientWithVehicleRoundtrip(t *testing.T) {
	repo := NewClientRepository(testPool)
	ctx := context.Background()

	c := &domain.Client{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Email:     "it.registration.roundtrip@example.com",
		FullName:  "IT Registration",
		Phone:     "+79990000010",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	v := &domain.ClientVehicle{
		ID:        uuid.New(),
		VehicleID: uuid.New(),
		VIN:       "ITVINROUNDTRIP00001",
		Make:      "Toyota",
		Model:     "Camry",
		Year:      2021,
		AddedAt:   time.Now().UTC(),
	}

	if err := repo.CreateClientWithVehicle(ctx, c, v); err != nil {
		t.Fatalf("CreateClientWithVehicle: %v", err)
	}
	if v.ClientID != c.ID {
		t.Fatalf("vehicle client_id: got %v want %v", v.ClientID, c.ID)
	}

	got, err := repo.GetByUserID(ctx, c.UserID)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if got.ID != c.ID || got.Email != c.Email || got.FullName != c.FullName || got.Phone != c.Phone {
		t.Fatalf("GetByUserID mismatch: %+v", got)
	}

	vehicles, err := repo.ListVehiclesByClientID(ctx, c.ID)
	if err != nil {
		t.Fatalf("ListVehiclesByClientID: %v", err)
	}
	if len(vehicles) != 1 {
		t.Fatalf("vehicles len: got %d want 1", len(vehicles))
	}
	gotV := vehicles[0]
	if gotV.ID != v.ID || gotV.VIN != v.VIN || gotV.Make != v.Make || gotV.Year != v.Year {
		t.Fatalf("vehicle mismatch: %+v", gotV)
	}

	emailExists, err := repo.EmailExists(ctx, c.Email)
	if err != nil {
		t.Fatalf("EmailExists: %v", err)
	}
	if !emailExists {
		t.Fatal("EmailExists: want true")
	}
	vinLinked, err := repo.VINLinked(ctx, v.VIN)
	if err != nil {
		t.Fatalf("VINLinked: %v", err)
	}
	if !vinLinked {
		t.Fatal("VINLinked: want true")
	}

	if _, err := repo.GetByUserID(ctx, uuid.New()); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("GetByUserID missing: got %v want pgx.ErrNoRows", err)
	}
	notExists, err := repo.EmailExists(ctx, "it.registration.missing@example.com")
	if err != nil {
		t.Fatalf("EmailExists missing: %v", err)
	}
	if notExists {
		t.Fatal("EmailExists missing: want false")
	}
	notLinked, err := repo.VINLinked(ctx, "ITVINNOTEXIST00001")
	if err != nil {
		t.Fatalf("VINLinked missing: %v", err)
	}
	if notLinked {
		t.Fatal("VINLinked missing: want false")
	}
}

func TestClientRepository_DuplicateVIN(t *testing.T) {
	repo := NewClientRepository(testPool)
	ctx := context.Background()

	vin := "ITVINDUP0000000001"

	a := &domain.Client{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Email:     "it.dup.vin.a@example.com",
		FullName:  "A",
		Phone:     "+79990000011",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	va := &domain.ClientVehicle{ID: uuid.New(), VehicleID: uuid.New(), VIN: vin, Make: "Toyota", Model: "Camry", Year: 2020, AddedAt: time.Now().UTC()}
	if err := repo.CreateClientWithVehicle(ctx, a, va); err != nil {
		t.Fatalf("first CreateClientWithVehicle: %v", err)
	}

	b := &domain.Client{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Email:     "it.dup.vin.b@example.com",
		FullName:  "B",
		Phone:     "+79990000012",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	vb := &domain.ClientVehicle{ID: uuid.New(), VehicleID: uuid.New(), VIN: vin, Make: "Honda", Model: "Accord", Year: 2022, AddedAt: time.Now().UTC()}
	err := repo.CreateClientWithVehicle(ctx, b, vb)
	if !errors.Is(err, ErrVINAlreadyLinked) {
		t.Fatalf("duplicate VIN: got %v want ErrVINAlreadyLinked", err)
	}
	if _, err := repo.GetByUserID(ctx, b.UserID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("rolled back client should be missing: got %v want pgx.ErrNoRows", err)
	}

	v2 := &domain.ClientVehicle{ID: uuid.New(), VehicleID: uuid.New(), VIN: vin, Make: "Ford", Model: "Focus", Year: 2019, AddedAt: time.Now().UTC()}
	if err := repo.AddVehicle(ctx, a.ID, v2); !errors.Is(err, ErrVINAlreadyLinked) {
		t.Fatalf("AddVehicle duplicate VIN: got %v want ErrVINAlreadyLinked", err)
	}
}

func TestClientRepository_AddVehicle(t *testing.T) {
	repo := NewClientRepository(testPool)
	ctx := context.Background()

	c := &domain.Client{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Email:     "it.add.vehicle@example.com",
		FullName:  "AddVehicle",
		Phone:     "+79990000013",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	first := &domain.ClientVehicle{ID: uuid.New(), VehicleID: uuid.New(), VIN: "ITVINADDFIRST00001", Make: "Toyota", Model: "Camry", Year: 2021, AddedAt: time.Now().UTC()}
	if err := repo.CreateClientWithVehicle(ctx, c, first); err != nil {
		t.Fatalf("CreateClientWithVehicle: %v", err)
	}

	second := &domain.ClientVehicle{VehicleID: uuid.New(), VIN: "ITVINADDSECOND00001", Make: "Honda", Model: "Civic", Year: 2020}
	if err := repo.AddVehicle(ctx, c.ID, second); err != nil {
		t.Fatalf("AddVehicle: %v", err)
	}
	if second.ID == uuid.Nil {
		t.Fatal("AddVehicle: ID not populated")
	}
	if second.ClientID != c.ID {
		t.Fatalf("AddVehicle client_id: got %v want %v", second.ClientID, c.ID)
	}
	if second.AddedAt.IsZero() {
		t.Fatal("AddVehicle: AddedAt not populated")
	}

	vehicles, err := repo.ListVehiclesByClientID(ctx, c.ID)
	if err != nil {
		t.Fatalf("ListVehiclesByClientID: %v", err)
	}
	if len(vehicles) != 2 {
		t.Fatalf("vehicles len: got %d want 2", len(vehicles))
	}
}
