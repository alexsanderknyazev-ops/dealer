package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/employees/internal/domain"
)

type memEmployeeRepo struct {
	byID   map[uuid.UUID]*domain.Employee
	byUser map[uuid.UUID]*domain.Employee
	err    error
}

func (m *memEmployeeRepo) Create(_ context.Context, e *domain.Employee) error {
	if m.err != nil {
		return m.err
	}
	if m.byID == nil {
		m.byID = make(map[uuid.UUID]*domain.Employee)
	}
	if m.byUser == nil {
		m.byUser = make(map[uuid.UUID]*domain.Employee)
	}
	cp := *e
	m.byID[e.ID] = &cp
	if e.UserID != nil {
		m.byUser[*e.UserID] = &cp
	}
	return nil
}

func (m *memEmployeeRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Employee, error) {
	if m.err != nil {
		return nil, m.err
	}
	e, ok := m.byID[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *e
	return &cp, nil
}

func (m *memEmployeeRepo) GetByUserID(_ context.Context, userID uuid.UUID) (*domain.Employee, error) {
	if m.err != nil {
		return nil, m.err
	}
	e, ok := m.byUser[userID]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *e
	return &cp, nil
}

func (m *memEmployeeRepo) ResolveRef(_ context.Context, ref uuid.UUID) (*domain.Employee, error) {
	if e, err := m.GetByID(context.Background(), ref); err == nil {
		return e, nil
	}
	return m.GetByUserID(context.Background(), ref)
}

func (m *memEmployeeRepo) List(_ context.Context, _, _ int32, _, _ string, _ bool) ([]*domain.Employee, int32, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	var out []*domain.Employee
	for _, e := range m.byID {
		cp := *e
		out = append(out, &cp)
	}
	return out, int32(len(out)), nil
}

func (m *memEmployeeRepo) Update(_ context.Context, e *domain.Employee) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.byID[e.ID]; !ok {
		return pgx.ErrNoRows
	}
	cp := *e
	m.byID[e.ID] = &cp
	if e.UserID != nil {
		m.byUser[*e.UserID] = &cp
	}
	return nil
}

func (m *memEmployeeRepo) Delete(_ context.Context, id uuid.UUID) error {
	if m.err != nil {
		return m.err
	}
	delete(m.byID, id)
	return nil
}

func TestEmployeeService_Create(t *testing.T) {
	s := NewEmployeeService(&memEmployeeRepo{byID: map[uuid.UUID]*domain.Employee{}})
	userID := uuid.New().String()
	e, err := s.Create(context.Background(), userID, " QA Name ", "master", "СТО", "+7900", true)
	if err != nil {
		t.Fatal(err)
	}
	if e.FullName != "QA Name" || e.Position != "master" || !e.Active {
		t.Fatalf("%+v", e)
	}
	if e.UserID == nil || e.UserID.String() != userID {
		t.Fatalf("user_id=%v", e.UserID)
	}
}

func TestEmployeeService_GetByUserID(t *testing.T) {
	repo := &memEmployeeRepo{byID: map[uuid.UUID]*domain.Employee{}}
	s := NewEmployeeService(repo)
	userID := uuid.New()
	e, _ := s.Create(context.Background(), userID.String(), "Name", "sales", "", "", true)

	found, err := s.GetByUserID(context.Background(), userID.String())
	if err != nil || found.ID != e.ID {
		t.Fatalf("err=%v id=%v", err, found)
	}
}

func TestEmployeeService_ResolveRef_ByUserID(t *testing.T) {
	repo := &memEmployeeRepo{byID: map[uuid.UUID]*domain.Employee{}}
	s := NewEmployeeService(repo)
	userID := uuid.New()
	e, _ := s.Create(context.Background(), userID.String(), "Name", "sales", "", "", true)

	found, err := s.ResolveRef(context.Background(), userID.String())
	if err != nil || found.ID != e.ID {
		t.Fatalf("err=%v", err)
	}
}

func TestEmployeeService_Get_NotFound(t *testing.T) {
	s := NewEmployeeService(&memEmployeeRepo{byID: map[uuid.UUID]*domain.Employee{}})
	_, err := s.Get(context.Background(), uuid.New().String())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestEmployeeService_Update(t *testing.T) {
	repo := &memEmployeeRepo{byID: map[uuid.UUID]*domain.Employee{}}
	s := NewEmployeeService(repo)
	e, _ := s.Create(context.Background(), "", "Old", "sales", "", "", true)

	name := "New"
	updated, err := s.Update(context.Background(), e.ID.String(), nil, &name, nil, nil, nil, nil)
	if err != nil || updated.FullName != "New" {
		t.Fatalf("err=%v name=%q", err, updated.FullName)
	}
}

func TestEmployeeService_List_DefaultLimit(t *testing.T) {
	s := NewEmployeeService(&memEmployeeRepo{byID: map[uuid.UUID]*domain.Employee{}})
	_, _, err := s.List(context.Background(), 0, 0, "", "", false)
	if err != nil {
		t.Fatal(err)
	}
}
