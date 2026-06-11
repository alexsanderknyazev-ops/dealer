package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"testing"

	"github.com/dealer/dealer/services/deals/internal/domain"
)

type memDealRepo struct {
	byID      map[uuid.UUID]*domain.Deal
	err       error
	updateErr error
}

type memRefs struct {
	err    error
	custOK *bool
	vehOK  *bool
}

func (m *memRefs) CustomerExists(_ context.Context, _ uuid.UUID) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	if m.custOK == nil {
		return true, nil
	}
	return *m.custOK, nil
}

func (m *memRefs) VehicleExists(_ context.Context, _ uuid.UUID) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	if m.vehOK == nil {
		return true, nil
	}
	return *m.vehOK, nil
}

func (m *memDealRepo) Create(_ context.Context, d *domain.Deal) error {
	if m.err != nil {
		return m.err
	}
	if m.byID == nil {
		m.byID = make(map[uuid.UUID]*domain.Deal)
	}
	cp := *d
	m.byID[d.ID] = &cp
	return nil
}

func (m *memDealRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Deal, error) {
	if m.err != nil {
		return nil, m.err
	}
	d, ok := m.byID[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *d
	return &cp, nil
}

func (m *memDealRepo) List(_ context.Context, _, _ int32, _, _ string) ([]*domain.Deal, int32, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	var out []*domain.Deal
	for _, d := range m.byID {
		cp := *d
		out = append(out, &cp)
	}
	return out, int32(len(out)), nil
}

func (m *memDealRepo) Update(_ context.Context, d *domain.Deal) error {
	if m.err != nil {
		return m.err
	}
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.byID[d.ID]; !ok {
		return pgx.ErrNoRows
	}
	cp := *d
	m.byID[d.ID] = &cp
	return nil
}

func (m *memDealRepo) Delete(_ context.Context, id uuid.UUID) error {
	if m.err != nil {
		return m.err
	}
	delete(m.byID, id)
	return nil
}

func TestDealService_Create_EmptyAmountDefaultsToZero(t *testing.T) {
	r := &memDealRepo{byID: map[uuid.UUID]*domain.Deal{}}
	s := NewDealService(r, nil)
	cid, vid := uuid.New(), uuid.New()
	d, err := s.Create(context.Background(), CreateDealInput{CustomerID: cid.String(), VehicleID: vid.String()})
	if err != nil || d.Amount != "0" {
		t.Fatalf("%v %+v", err, d)
	}
}

func TestDealService_Create_DefaultStage(t *testing.T) {
	r := &memDealRepo{byID: map[uuid.UUID]*domain.Deal{}}
	s := NewDealService(r, nil)
	cid, vid := uuid.New(), uuid.New()
	d, err := s.Create(context.Background(), CreateDealInput{CustomerID: cid.String(), VehicleID: vid.String(), Amount: "100"})
	if err != nil || d.Stage != "draft" {
		t.Fatalf("%v %+v", err, d)
	}
}

func TestDealService_Create_InvalidCustomer(t *testing.T) {
	s := NewDealService(&memDealRepo{}, nil)
	_, err := s.Create(context.Background(), CreateDealInput{CustomerID: "bad", VehicleID: uuid.New().String()})
	if err == nil || err.Error() != "invalid customer_id" {
		t.Fatalf("%v", err)
	}
}

func TestDealService_Create_ReferenceIntegrity(t *testing.T) {
	customerMissing := false
	vehicleMissing := false
	s1 := NewDealService(&memDealRepo{}, &memRefs{custOK: &customerMissing})
	_, err := s1.Create(context.Background(), CreateDealInput{CustomerID: uuid.New().String(), VehicleID: uuid.New().String()})
	if !errors.Is(err, ErrCustomerNotFound) {
		t.Fatalf("want ErrCustomerNotFound, got %v", err)
	}

	s2 := NewDealService(&memDealRepo{}, &memRefs{vehOK: &vehicleMissing})
	_, err = s2.Create(context.Background(), CreateDealInput{CustomerID: uuid.New().String(), VehicleID: uuid.New().String()})
	if !errors.Is(err, ErrVehicleNotFound) {
		t.Fatalf("want ErrVehicleNotFound, got %v", err)
	}
}

func TestDealService_Get_NotFound(t *testing.T) {
	s := NewDealService(&memDealRepo{byID: map[uuid.UUID]*domain.Deal{}}, nil)
	_, err := s.Get(context.Background(), uuid.New().String())
	if err != ErrNotFound {
		t.Fatalf("%v", err)
	}
	_, err = s.Get(context.Background(), "x")
	if err != ErrNotFound {
		t.Fatalf("%v", err)
	}
}

func TestDealService_Update_AssignedEmptyClears(t *testing.T) {
	r := &memDealRepo{byID: map[uuid.UUID]*domain.Deal{}}
	s := NewDealService(r, nil)
	cid, vid := uuid.New(), uuid.New()
	a := uuid.New()
	d, _ := s.Create(context.Background(), CreateDealInput{CustomerID: cid.String(), VehicleID: vid.String(), Amount: "1", Stage: "open", AssignedTo: a.String()})
	empty := ""
	d2, err := s.Update(context.Background(), d.ID.String(), UpdateDealInput{AssignedTo: &empty})
	if err != nil || d2.AssignedTo != nil {
		t.Fatalf("%v %+v", err, d2.AssignedTo)
	}
}

func TestDealService_List_DefaultLimit(t *testing.T) {
	s := NewDealService(&memDealRepo{byID: map[uuid.UUID]*domain.Deal{}}, nil)
	_, _, err := s.List(context.Background(), 0, 0, "", "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDealService_Create_Err(t *testing.T) {
	s := NewDealService(&memDealRepo{byID: map[uuid.UUID]*domain.Deal{}, err: errors.New("db")}, nil)
	_, err := s.Create(context.Background(), CreateDealInput{CustomerID: uuid.New().String(), VehicleID: uuid.New().String()})
	if err == nil {
		t.Fatal("want err")
	}
}

func TestDealService_Delete_NotFound(t *testing.T) {
	s := NewDealService(&memDealRepo{byID: map[uuid.UUID]*domain.Deal{}}, nil)
	err := s.Delete(context.Background(), uuid.New().String())
	if err != nil {
		t.Fatal(err)
	}
	err = s.Delete(context.Background(), "bad")
	if err != ErrNotFound {
		t.Fatalf("%v", err)
	}
}

func TestDealService_Update_GetExisting(t *testing.T) {
	r := &memDealRepo{byID: map[uuid.UUID]*domain.Deal{}}
	s := NewDealService(r, nil)
	cid, vid := uuid.New(), uuid.New()
	d, _ := s.Create(context.Background(), CreateDealInput{CustomerID: cid.String(), VehicleID: vid.String(), Amount: "10", Stage: "x"})
	amt := "20"
	d2, err := s.Update(context.Background(), d.ID.String(), UpdateDealInput{Amount: &amt})
	if err != nil || d2.Amount != "20" {
		t.Fatalf("%v", err)
	}
}

func TestDealService_Get_DBErr(t *testing.T) {
	s := NewDealService(&memDealRepo{byID: map[uuid.UUID]*domain.Deal{}, err: errors.New("db")}, nil)
	_, err := s.Get(context.Background(), uuid.New().String())
	if err == nil || errors.Is(err, ErrNotFound) {
		t.Fatalf("%v", err)
	}
}

func TestDealService_Update_NotFound(t *testing.T) {
	s := NewDealService(&memDealRepo{byID: map[uuid.UUID]*domain.Deal{}}, nil)
	x := "x"
	_, err := s.Update(context.Background(), uuid.New().String(), UpdateDealInput{Amount: &x})
	if err != ErrNotFound {
		t.Fatalf("%v", err)
	}
}

func TestDealService_Update_UpdateFails(t *testing.T) {
	r := &memDealRepo{byID: map[uuid.UUID]*domain.Deal{}, updateErr: errors.New("db")}
	s := NewDealService(r, nil)
	cid, vid := uuid.New(), uuid.New()
	d, _ := s.Create(context.Background(), CreateDealInput{CustomerID: cid.String(), VehicleID: vid.String(), Amount: "1", Stage: "x"})
	x := "2"
	_, err := s.Update(context.Background(), d.ID.String(), UpdateDealInput{Amount: &x})
	if err == nil {
		t.Fatal("want err")
	}
}
