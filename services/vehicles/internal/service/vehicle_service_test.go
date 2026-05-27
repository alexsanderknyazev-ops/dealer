package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/vehicles/internal/domain"
)

type memVehicleRepo struct {
	byID      map[uuid.UUID]*domain.Vehicle
	err       error
	updateErr error
}

func (m *memVehicleRepo) Create(_ context.Context, v *domain.Vehicle) error {
	if m.err != nil {
		return m.err
	}
	if m.byID == nil {
		m.byID = make(map[uuid.UUID]*domain.Vehicle)
	}
	cp := *v
	m.byID[v.ID] = &cp
	return nil
}

func (m *memVehicleRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Vehicle, error) {
	if m.err != nil {
		return nil, m.err
	}
	v, ok := m.byID[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *v
	return &cp, nil
}

func (m *memVehicleRepo) List(_ context.Context, _ domain.VehicleListFilter) ([]*domain.Vehicle, int32, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	var out []*domain.Vehicle
	for _, v := range m.byID {
		cp := *v
		out = append(out, &cp)
	}
	return out, int32(len(out)), nil
}

func (m *memVehicleRepo) Update(_ context.Context, v *domain.Vehicle) error {
	if m.err != nil {
		return m.err
	}
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.byID[v.ID]; !ok {
		return pgx.ErrNoRows
	}
	cp := *v
	m.byID[v.ID] = &cp
	return nil
}

func (m *memVehicleRepo) Delete(_ context.Context, id uuid.UUID) error {
	if m.err != nil {
		return m.err
	}
	delete(m.byID, id)
	return nil
}

func TestVehicleService_Create_DefaultStatus(t *testing.T) {
	r := &memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}}
	s := NewVehicleService(r)
	v, err := s.Create(context.Background(), CreateVehicleInput{VIN: "VIN1", Make: "M", Model: "X", Year: 2020, MileageKm: 0, Price: "100"})
	if err != nil || v.Status != "available" {
		t.Fatalf("%v %+v", err, v)
	}
}

func TestVehicleService_Get_NotFound(t *testing.T) {
	s := NewVehicleService(&memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}})
	_, err := s.Get(context.Background(), uuid.New().String())
	if err != ErrNotFound {
		t.Fatalf("%v", err)
	}
	_, err = s.Get(context.Background(), "bad-id")
	if err != ErrNotFound {
		t.Fatalf("%v", err)
	}
}

func TestVehicleService_List_DefaultLimit(t *testing.T) {
	s := NewVehicleService(&memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}})
	_, _, err := s.List(context.Background(), domain.VehicleListFilter{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestVehicleService_Update_Delete(t *testing.T) {
	r := &memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}}
	s := NewVehicleService(r)
	v, _ := s.Create(context.Background(), CreateVehicleInput{VIN: "V", Make: "mk", Model: "md", Year: 2021, MileageKm: 1, Price: "1", Status: "sold", Color: "c", Notes: "n"})
	nm := "newmake"
	upd, err := s.Update(context.Background(), v.ID.String(), UpdateVehicleInput{Make: &nm})
	if err != nil || upd.Make != "newmake" {
		t.Fatalf("%v", err)
	}
	if err := s.Delete(context.Background(), v.ID.String()); err != nil {
		t.Fatal(err)
	}
}

func TestVehicleService_Update_ClearBrand(t *testing.T) {
	r := &memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}}
	s := NewVehicleService(r)
	bid := uuid.New()
	v, _ := s.Create(context.Background(), CreateVehicleInput{VIN: "V2", Make: "m", Model: "m", Year: 2022, MileageKm: 0, Price: "0", Status: "a", BrandID: &bid})
	upd, err := s.Update(context.Background(), v.ID.String(), UpdateVehicleInput{ClearBrand: true})
	if err != nil || upd.BrandID != nil {
		t.Fatalf("%v %+v", err, upd.BrandID)
	}
}

func TestVehicleService_Create_Err(t *testing.T) {
	s := NewVehicleService(&memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}, err: errors.New("db")})
	_, err := s.Create(context.Background(), CreateVehicleInput{VIN: "x"})
	if err == nil {
		t.Fatal("want err")
	}
}

func TestVehicleService_Get_DBErr(t *testing.T) {
	s := NewVehicleService(&memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}, err: errors.New("db")})
	_, err := s.Get(context.Background(), uuid.New().String())
	if err == nil || errors.Is(err, ErrNotFound) {
		t.Fatalf("%v", err)
	}
}

func TestVehicleService_List_NormalizesLimit(t *testing.T) {
	r := &memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}}
	s := NewVehicleService(r)
	_, total, err := s.List(context.Background(), domain.VehicleListFilter{Limit: 500})
	if err != nil || total != 0 {
		t.Fatalf("%v %d", err, total)
	}
}

func TestVehicleService_Update_RepoFails(t *testing.T) {
	r := &memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}, updateErr: errors.New("db")}
	s := NewVehicleService(r)
	v, _ := s.Create(context.Background(), CreateVehicleInput{VIN: "V", Make: "m", Model: "m", Year: 2020, MileageKm: 0, Price: "1", Status: "a"})
	mk := "z"
	_, err := s.Update(context.Background(), v.ID.String(), UpdateVehicleInput{Make: &mk})
	if err == nil {
		t.Fatal("want err")
	}
}
