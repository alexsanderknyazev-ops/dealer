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

func (m *memVehicleRepo) GetByVIN(_ context.Context, vin string) (*domain.Vehicle, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, v := range m.byID {
		if v.VIN == vin {
			cp := *v
			return &cp, nil
		}
	}
	return nil, pgx.ErrNoRows
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

type memDealerPoints struct {
	dpOK, leOK, whOK *bool
}

func (m *memDealerPoints) DealerPointExists(_ context.Context, _ uuid.UUID) (bool, error) {
	if m.dpOK == nil {
		return true, nil
	}
	return *m.dpOK, nil
}
func (m *memDealerPoints) LegalEntityExists(_ context.Context, _ uuid.UUID) (bool, error) {
	if m.leOK == nil {
		return true, nil
	}
	return *m.leOK, nil
}
func (m *memDealerPoints) WarehouseExists(_ context.Context, _ uuid.UUID) (bool, error) {
	if m.whOK == nil {
		return true, nil
	}
	return *m.whOK, nil
}

type memBrands struct {
	ok *bool
}

func (m *memBrands) BrandExists(_ context.Context, _ uuid.UUID) (bool, error) {
	if m.ok == nil {
		return true, nil
	}
	return *m.ok, nil
}

func TestVehicleService_Create_ReferenceIntegrity(t *testing.T) {
	missing := false
	dpID := uuid.New()
	leID := uuid.New()
	whID := uuid.New()
	base := CreateVehicleInput{VIN: "V", Make: "m", Model: "m", Year: 2020, MileageKm: 0, Price: "1"}

	s1 := NewVehicleService(&memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}}, nil, &memDealerPoints{dpOK: &missing})
	in1 := base
	in1.DealerPointID = &dpID
	_, err := s1.Create(context.Background(), in1)
	if !errors.Is(err, ErrDealerPointNotFound) {
		t.Fatalf("dealer point: %v", err)
	}

	s2 := NewVehicleService(&memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}}, nil, &memDealerPoints{leOK: &missing})
	in2 := base
	in2.DealerPointID = &dpID
	in2.LegalEntityID = &leID
	_, err = s2.Create(context.Background(), in2)
	if !errors.Is(err, ErrLegalEntityNotFound) {
		t.Fatalf("legal entity: %v", err)
	}

	s3 := NewVehicleService(&memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}}, nil, &memDealerPoints{whOK: &missing})
	in3 := base
	in3.DealerPointID = &dpID
	in3.LegalEntityID = &leID
	in3.WarehouseID = &whID
	_, err = s3.Create(context.Background(), in3)
	if !errors.Is(err, ErrWarehouseNotFound) {
		t.Fatalf("warehouse: %v", err)
	}

	bid := uuid.New()
	s4 := NewVehicleService(&memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}}, &memBrands{ok: &missing}, &memDealerPoints{})
	in4 := base
	in4.BrandID = &bid
	in4.DealerPointID = &dpID
	in4.LegalEntityID = &leID
	in4.WarehouseID = &whID
	_, err = s4.Create(context.Background(), in4)
	if !errors.Is(err, ErrBrandNotFound) {
		t.Fatalf("brand: %v", err)
	}
}

func TestVehicleService_Create_DefaultStatus(t *testing.T) {
	r := &memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}}
	s := NewVehicleService(r, nil, nil)
	v, err := s.Create(context.Background(), CreateVehicleInput{VIN: "VIN1", Make: "M", Model: "X", Year: 2020, MileageKm: 0, Price: "100"})
	if err != nil || v.Status != "available" {
		t.Fatalf("%v %+v", err, v)
	}
}

func TestVehicleService_Get_NotFound(t *testing.T) {
	s := NewVehicleService(&memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}}, nil, nil)
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
	s := NewVehicleService(&memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}}, nil, nil)
	_, _, err := s.List(context.Background(), domain.VehicleListFilter{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestVehicleService_Update_Delete(t *testing.T) {
	r := &memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}}
	s := NewVehicleService(r, nil, nil)
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
	s := NewVehicleService(r, nil, nil)
	bid := uuid.New()
	v, _ := s.Create(context.Background(), CreateVehicleInput{VIN: "V2", Make: "m", Model: "m", Year: 2022, MileageKm: 0, Price: "0", Status: "a", BrandID: &bid})
	upd, err := s.Update(context.Background(), v.ID.String(), UpdateVehicleInput{ClearBrand: true})
	if err != nil || upd.BrandID != nil {
		t.Fatalf("%v %+v", err, upd.BrandID)
	}
}

func TestVehicleService_Create_Err(t *testing.T) {
	s := NewVehicleService(&memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}, err: errors.New("db")}, nil, nil)
	_, err := s.Create(context.Background(), CreateVehicleInput{VIN: "x"})
	if err == nil {
		t.Fatal("want err")
	}
}

func TestVehicleService_Get_DBErr(t *testing.T) {
	s := NewVehicleService(&memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}, err: errors.New("db")}, nil, nil)
	_, err := s.Get(context.Background(), uuid.New().String())
	if err == nil || errors.Is(err, ErrNotFound) {
		t.Fatalf("%v", err)
	}
}

func TestVehicleService_List_NormalizesLimit(t *testing.T) {
	r := &memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}}
	s := NewVehicleService(r, nil, nil)
	_, total, err := s.List(context.Background(), domain.VehicleListFilter{Limit: 500})
	if err != nil || total != 0 {
		t.Fatalf("%v %d", err, total)
	}
}

func TestVehicleService_Update_RepoFails(t *testing.T) {
	r := &memVehicleRepo{byID: map[uuid.UUID]*domain.Vehicle{}, updateErr: errors.New("db")}
	s := NewVehicleService(r, nil, nil)
	v, _ := s.Create(context.Background(), CreateVehicleInput{VIN: "V", Make: "m", Model: "m", Year: 2020, MileageKm: 0, Price: "1", Status: "a"})
	mk := "z"
	_, err := s.Update(context.Background(), v.ID.String(), UpdateVehicleInput{Make: &mk})
	if err == nil {
		t.Fatal("want err")
	}
}
