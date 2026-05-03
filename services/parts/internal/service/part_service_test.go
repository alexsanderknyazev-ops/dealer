package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

type fakePartRepo struct {
	parts map[uuid.UUID]*domain.Part
	err   error
}

func (f *fakePartRepo) Create(_ context.Context, p *domain.Part) error {
	if f.err != nil {
		return f.err
	}
	if f.parts == nil {
		f.parts = make(map[uuid.UUID]*domain.Part)
	}
	cp := *p
	f.parts[p.ID] = &cp
	return nil
}

func (f *fakePartRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Part, error) {
	if f.err != nil {
		return nil, f.err
	}
	p, ok := f.parts[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *p
	return &cp, nil
}

func (f *fakePartRepo) List(_ context.Context, _ domain.PartListFilter) ([]*domain.Part, int32, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	var out []*domain.Part
	for _, p := range f.parts {
		cp := *p
		out = append(out, &cp)
	}
	return out, int32(len(out)), nil
}

func (f *fakePartRepo) Update(_ context.Context, p *domain.Part) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.parts[p.ID]; !ok {
		return pgx.ErrNoRows
	}
	cp := *p
	f.parts[p.ID] = &cp
	return nil
}

func (f *fakePartRepo) Delete(_ context.Context, id uuid.UUID) error {
	if f.err != nil {
		return f.err
	}
	delete(f.parts, id)
	return nil
}

type fakeStock struct {
	repo *fakePartRepo
}

func (fakeStock) ListByPart(context.Context, uuid.UUID) ([]*domain.PartStock, error) {
	return nil, nil
}

func (f *fakeStock) Upsert(_ context.Context, partID, _ uuid.UUID, quantity int32) error {
	if p, ok := f.repo.parts[partID]; ok {
		p.Quantity = quantity
	}
	return nil
}

func (f *fakeStock) ReplaceForPart(_ context.Context, partID uuid.UUID, rows []struct {
	WarehouseID uuid.UUID
	Quantity    int32
}) error {
	var sum int32
	for _, r := range rows {
		sum += r.Quantity
	}
	if p, ok := f.repo.parts[partID]; ok {
		p.Quantity = sum
	}
	return nil
}

type fakeFolderRepo struct {
	folders map[uuid.UUID]*domain.PartFolder
}

func (f *fakeFolderRepo) Create(_ context.Context, folder *domain.PartFolder) error {
	if f.folders == nil {
		f.folders = make(map[uuid.UUID]*domain.PartFolder)
	}
	cp := *folder
	f.folders[folder.ID] = &cp
	return nil
}

func (f *fakeFolderRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.PartFolder, error) {
	x, ok := f.folders[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *x
	return &cp, nil
}

func (f *fakeFolderRepo) ListByParent(context.Context, *uuid.UUID) ([]*domain.PartFolder, error) {
	return nil, nil
}

func (f *fakeFolderRepo) Update(_ context.Context, folder *domain.PartFolder) error {
	if _, ok := f.folders[folder.ID]; !ok {
		return pgx.ErrNoRows
	}
	cp := *folder
	f.folders[folder.ID] = &cp
	return nil
}

func (f *fakeFolderRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(f.folders, id)
	return nil
}

func TestPartService_Create_DefaultUnit(t *testing.T) {
	pr := &fakePartRepo{parts: map[uuid.UUID]*domain.Part{}}
	s := NewPartService(pr, &fakeFolderRepo{}, &fakeStock{repo: pr})
	p, err := s.Create(context.Background(), "SKU1", "N", "cat", nil, nil, nil, nil, nil, 0, "", "10", "", "", nil)
	if err != nil || p.Unit != "шт" {
		t.Fatalf("%v %+v", err, p)
	}
}

func TestPartService_Get_NotFound(t *testing.T) {
	pr := &fakePartRepo{parts: map[uuid.UUID]*domain.Part{}}
	s := NewPartService(pr, &fakeFolderRepo{}, &fakeStock{repo: pr})
	_, err := s.Get(context.Background(), uuid.New().String())
	if err != ErrNotFound {
		t.Fatalf("%v", err)
	}
	_, err = s.Get(context.Background(), "bad")
	if err != ErrNotFound {
		t.Fatalf("%v", err)
	}
}

func TestPartService_Create_WithWarehouseQty(t *testing.T) {
	pr := &fakePartRepo{parts: map[uuid.UUID]*domain.Part{}}
	st := &fakeStock{repo: pr}
	s := NewPartService(pr, &fakeFolderRepo{}, st)
	wid := uuid.New()
	p, err := s.Create(context.Background(), "S2", "N", "c", nil, nil, nil, nil, &wid, 5, "", "", "", "", nil)
	if err != nil || p.Quantity != 5 {
		t.Fatalf("%v q=%d", err, p.Quantity)
	}
}

func TestPartService_FolderCRUD(t *testing.T) {
	pr := &fakePartRepo{parts: map[uuid.UUID]*domain.Part{}}
	fr := &fakeFolderRepo{folders: map[uuid.UUID]*domain.PartFolder{}}
	s := NewPartService(pr, fr, &fakeStock{repo: pr})
	f, err := s.CreateFolder(context.Background(), "root", nil)
	if err != nil || f.Name != "root" {
		t.Fatal(err)
	}
	g, err := s.GetFolder(context.Background(), f.ID.String())
	if err != nil || g.ID != f.ID {
		t.Fatal(err)
	}
}

func TestPartService_Update_Delete_ListStock(t *testing.T) {
	pr := &fakePartRepo{parts: map[uuid.UUID]*domain.Part{}}
	st := &fakeStock{repo: pr}
	s := NewPartService(pr, &fakeFolderRepo{}, st)
	p, err := s.Create(context.Background(), "U1", "Part", "c", nil, nil, nil, nil, nil, 0, "шт", "1", "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	sku := "U2"
	upd, err := s.Update(context.Background(), p.ID.String(),
		&sku, nil, nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
	)
	if err != nil || upd.SKU != "U2" {
		t.Fatalf("%v", err)
	}
	if _, err := s.ListStock(context.Background(), p.ID.String()); err != nil {
		t.Fatal(err)
	}
	wid := uuid.New()
	if err := s.ReplaceStock(context.Background(), p.ID.String(), []StockRow{{WarehouseID: wid, Quantity: 3}}); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(context.Background(), p.ID.String()); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(context.Background(), "not-uuid"); err != ErrNotFound {
		t.Fatalf("%v", err)
	}
}
