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

func (m *memVehicleRepo) List(_ context.Context, _, _ int32, _, _ string, _, _, _, _ *uuid.UUID) ([]*domain.Vehicle, int32, error) {
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
	v, err := s.Create(context.Background(), "VIN1", "M", "X", 2020, 0, "100", "", "", "", nil, nil, nil, nil)
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
	_, _, err := s.List(context.Background(), 0, 0, "", "", nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVehicleService_Update_Delete(t *testing.T) {
	r := &memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}}
	s := NewVehicleService(r)
	v, _ := s.Create(context.Background(), "V", "mk", "md", 2021, 1, "1", "sold", "c", "n", nil, nil, nil, nil)
	nm := "newmake"
	upd, err := s.Update(context.Background(), v.ID.String(), nil, &nm, nil, nil, nil, nil, nil, nil, nil, nil, false, nil, nil, nil, false, false, false)
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
	v, _ := s.Create(context.Background(), "V2", "m", "m", 2022, 0, "0", "a", "", "", &bid, nil, nil, nil)
	upd, err := s.Update(context.Background(), v.ID.String(), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, true, nil, nil, nil, false, false, false)
	if err != nil || upd.BrandID != nil {
		t.Fatalf("%v %+v", err, upd.BrandID)
	}
}

func TestVehicleService_Create_Err(t *testing.T) {
	s := NewVehicleService(&memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}, err: errors.New("db")})
	_, err := s.Create(context.Background(), "x", "", "", 0, 0, "", "", "", "", nil, nil, nil, nil)
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
	_, total, err := s.List(context.Background(), 500, 0, "", "", nil, nil, nil, nil)
	if err != nil || total != 0 {
		t.Fatalf("%v %d", err, total)
	}
}

func TestVehicleService_Update_RepoFails(t *testing.T) {
	r := &memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}, updateErr: errors.New("db")}
	s := NewVehicleService(r)
	v, _ := s.Create(context.Background(), "V", "m", "m", 2020, 0, "1", "a", "", "", nil, nil, nil, nil)
	mk := "z"
	_, err := s.Update(context.Background(), v.ID.String(),
		nil, &mk, nil, nil, nil,
		nil, nil, nil, nil,
		nil, false, nil, nil, nil, false, false, false,
	)
	if err == nil {
		t.Fatal("want err")
	}
}
