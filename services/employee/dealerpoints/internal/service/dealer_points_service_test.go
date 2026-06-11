package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/dealerpoints/internal/domain"
)

type memDP struct {
	byID map[uuid.UUID]*domain.DealerPoint
}

func (m *memDP) Create(_ context.Context, d *domain.DealerPoint) error {
	if m.byID == nil {
		m.byID = make(map[uuid.UUID]*domain.DealerPoint)
	}
	cp := *d
	m.byID[d.ID] = &cp
	return nil
}

func (m *memDP) GetByID(_ context.Context, id uuid.UUID) (*domain.DealerPoint, error) {
	d, ok := m.byID[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *d
	return &cp, nil
}

func (m *memDP) List(_ context.Context, _, _ int32, _ string) ([]*domain.DealerPoint, int32, error) {
	var out []*domain.DealerPoint
	for _, d := range m.byID {
		cp := *d
		out = append(out, &cp)
	}
	return out, int32(len(out)), nil
}

func (m *memDP) Update(_ context.Context, d *domain.DealerPoint) error {
	if _, ok := m.byID[d.ID]; !ok {
		return pgx.ErrNoRows
	}
	cp := *d
	m.byID[d.ID] = &cp
	return nil
}

func (m *memDP) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.byID, id)
	return nil
}

type memLE struct {
	byID map[uuid.UUID]*domain.LegalEntity
}

func (m *memLE) Create(_ context.Context, e *domain.LegalEntity) error {
	if m.byID == nil {
		m.byID = make(map[uuid.UUID]*domain.LegalEntity)
	}
	cp := *e
	m.byID[e.ID] = &cp
	return nil
}

func (m *memLE) GetByID(_ context.Context, id uuid.UUID) (*domain.LegalEntity, error) {
	e, ok := m.byID[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *e
	return &cp, nil
}

func (m *memLE) List(_ context.Context, _, _ int32, _ string) ([]*domain.LegalEntity, int32, error) {
	return nil, 0, nil
}

func (m *memLE) ListByDealerPoint(_ context.Context, _ uuid.UUID, _, _ int32) ([]*domain.LegalEntity, int32, error) {
	return nil, 0, nil
}

func (m *memLE) Update(_ context.Context, e *domain.LegalEntity) error {
	if _, ok := m.byID[e.ID]; !ok {
		return pgx.ErrNoRows
	}
	cp := *e
	m.byID[e.ID] = &cp
	return nil
}

func (m *memLE) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.byID, id)
	return nil
}

func (memLE) LinkToDealerPoint(context.Context, uuid.UUID, uuid.UUID) error   { return nil }
func (memLE) UnlinkFromDealerPoint(context.Context, uuid.UUID, uuid.UUID) error { return nil }

type memWH struct {
	byID map[uuid.UUID]*domain.Warehouse
}

func (m *memWH) Create(_ context.Context, w *domain.Warehouse) error {
	if m.byID == nil {
		m.byID = make(map[uuid.UUID]*domain.Warehouse)
	}
	cp := *w
	m.byID[w.ID] = &cp
	return nil
}

func (m *memWH) GetByID(_ context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	w, ok := m.byID[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *w
	return &cp, nil
}

func (m *memWH) List(_ context.Context, _, _ int32, _, _ *uuid.UUID, _ string) ([]*domain.Warehouse, int32, error) {
	return nil, 0, nil
}

func (m *memWH) Update(_ context.Context, w *domain.Warehouse) error {
	if _, ok := m.byID[w.ID]; !ok {
		return pgx.ErrNoRows
	}
	cp := *w
	m.byID[w.ID] = &cp
	return nil
}

func (m *memWH) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.byID, id)
	return nil
}

func TestDealerPoints_DP_CRUD(t *testing.T) {
	s := NewDealerPointsService(&memDP{byID: map[uuid.UUID]*domain.DealerPoint{}}, &memLE{}, &memWH{})
	d, err := s.CreateDealerPoint(context.Background(), "DP1", "addr")
	if err != nil || d.Name != "DP1" {
		t.Fatal(err)
	}
	g, err := s.GetDealerPoint(context.Background(), d.ID.String())
	if err != nil || g.ID != d.ID {
		t.Fatal(err)
	}
	_, err = s.GetDealerPoint(context.Background(), uuid.New().String())
	if err != ErrDealerPointNotFound {
		t.Fatalf("%v", err)
	}
}

func TestDealerPoints_Link_BadUUID(t *testing.T) {
	s := NewDealerPointsService(&memDP{}, &memLE{}, &memWH{})
	err := s.LinkLegalEntityToDealerPoint(context.Background(), "bad", uuid.New().String())
	if err != ErrDealerPointNotFound {
		t.Fatalf("%v", err)
	}
}

func TestDealerPoints_CreateWarehouse_DefaultType(t *testing.T) {
	dp := &memDP{byID: map[uuid.UUID]*domain.DealerPoint{}}
	le := &memLE{byID: map[uuid.UUID]*domain.LegalEntity{}}
	wh := &memWH{byID: map[uuid.UUID]*domain.Warehouse{}}
	s := NewDealerPointsService(dp, le, wh)
	d, _ := s.CreateDealerPoint(context.Background(), "x", "a")
	e, _ := s.CreateLegalEntity(context.Background(), "LE", "inn", "a")
	w, err := s.CreateWarehouse(context.Background(), d.ID.String(), e.ID.String(), "other", "W1")
	if err != nil || w.Type != "parts" {
		t.Fatalf("%v type=%s", err, w.Type)
	}
}
