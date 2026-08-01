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

func (f *fakeStock) ReplaceForPart(_ context.Context, partID uuid.UUID, rows []domain.PartWarehouseQty) error {
	var sum int32
	for _, r := range rows {
		sum += r.Quantity
	}
	if p, ok := f.repo.parts[partID]; ok {
		p.Quantity = sum
	}
	return nil
}

func (f *fakeStock) Add(_ context.Context, partID, _ uuid.UUID, quantity int32) error {
	if p, ok := f.repo.parts[partID]; ok {
		p.Quantity += quantity
	}
	return nil
}

func (f *fakeStock) Deduct(_ context.Context, partID, _ uuid.UUID, quantity int32) (int32, error) {
	if p, ok := f.repo.parts[partID]; ok {
		p.Quantity -= quantity
		return p.Quantity, nil
	}
	return 0, nil
}

func (f *fakeStock) GetQuantity(_ context.Context, partID, _ uuid.UUID) (int32, error) {
	if p, ok := f.repo.parts[partID]; ok {
		return p.Quantity, nil
	}
	return 0, nil
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

type memBrands struct {
	ok *bool
}

func (m *memBrands) BrandExists(_ context.Context, _ uuid.UUID) (bool, error) {
	if m.ok == nil {
		return true, nil
	}
	return *m.ok, nil
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

func TestPartService_Create_ReferenceIntegrity(t *testing.T) {
	missing := false
	pr := &fakePartRepo{parts: map[uuid.UUID]*domain.Part{}}
	fr := &fakeFolderRepo{folders: map[uuid.UUID]*domain.PartFolder{}}
	st := &fakeStock{repo: pr}
	bid := uuid.New()
	dpID := uuid.New()
	leID := uuid.New()
	whID := uuid.New()
	base := CreatePartInput{SKU: "S", Name: "N", Category: "c"}

	_, err := NewPartService(pr, fr, st, nil, nil, nil, &memBrands{ok: &missing}, &memDealerPoints{}, nil, nil, nil, nil).Create(
		context.Background(), func() CreatePartInput {
			in := base
			in.BrandID = &bid
			return in
		}(),
	)
	if err != ErrBrandNotFound {
		t.Fatalf("brand: %v", err)
	}

	_, err = NewPartService(pr, fr, st, nil, nil, nil, nil, &memDealerPoints{whOK: &missing}, nil, nil, nil, nil).Create(context.Background(), CreatePartInput{
		SKU: "S2", Name: "N", Category: "c", WarehouseID: &whID, Quantity: 1,
	})
	if err != ErrWarehouseNotFound {
		t.Fatalf("warehouse: %v", err)
	}

	_, err = NewPartService(pr, fr, st, nil, nil, nil, nil, &memDealerPoints{whOK: &missing}, nil, nil, nil, nil).Create(context.Background(), CreatePartInput{
		SKU: "S3", Name: "N", Category: "c",
		InitialStock: []StockRow{{WarehouseID: whID, Quantity: 2}},
	})
	if err != ErrWarehouseNotFound {
		t.Fatalf("stock warehouse: %v", err)
	}

	fid := uuid.New()
	_, err = NewPartService(pr, fr, st, nil, nil, nil, nil, &memDealerPoints{}, nil, nil, nil, nil).Create(context.Background(), CreatePartInput{
		SKU: "S4", Name: "N", Category: "c", FolderID: &fid,
	})
	if err != ErrFolderNotFound {
		t.Fatalf("folder: %v", err)
	}

	_, err = NewPartService(pr, fr, st, nil, nil, nil, nil, &memDealerPoints{dpOK: &missing}, nil, nil, nil, nil).Create(context.Background(), CreatePartInput{
		SKU: "S5", Name: "N", Category: "c", DealerPointID: &dpID,
	})
	if err != ErrDealerPointNotFound {
		t.Fatalf("dealer point: %v", err)
	}

	_, err = NewPartService(pr, fr, st, nil, nil, nil, nil, &memDealerPoints{leOK: &missing}, nil, nil, nil, nil).Create(context.Background(), CreatePartInput{
		SKU: "S6", Name: "N", Category: "c", DealerPointID: &dpID, LegalEntityID: &leID,
	})
	if err != ErrLegalEntityNotFound {
		t.Fatalf("legal entity: %v", err)
	}
}

func TestPartService_Create_DefaultUnit(t *testing.T) {
	pr := &fakePartRepo{parts: map[uuid.UUID]*domain.Part{}}
	s := NewPartService(pr, &fakeFolderRepo{}, &fakeStock{repo: pr}, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	p, err := s.Create(context.Background(), CreatePartInput{SKU: "SKU1", Name: "N", Category: "cat", Price: "10"})
	if err != nil || p.Unit != "шт" {
		t.Fatalf("%v %+v", err, p)
	}
}

func TestPartService_Get_NotFound(t *testing.T) {
	pr := &fakePartRepo{parts: map[uuid.UUID]*domain.Part{}}
	s := NewPartService(pr, &fakeFolderRepo{}, &fakeStock{repo: pr}, nil, nil, nil, nil, nil, nil, nil, nil, nil)
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
	s := NewPartService(pr, &fakeFolderRepo{}, st, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	wid := uuid.New()
	p, err := s.Create(context.Background(), CreatePartInput{SKU: "S2", Name: "N", Category: "c", WarehouseID: &wid, Quantity: 5})
	if err != nil || p.Quantity != 5 {
		t.Fatalf("%v q=%d", err, p.Quantity)
	}
}

func TestPartService_FolderCRUD(t *testing.T) {
	pr := &fakePartRepo{parts: map[uuid.UUID]*domain.Part{}}
	fr := &fakeFolderRepo{folders: map[uuid.UUID]*domain.PartFolder{}}
	s := NewPartService(pr, fr, &fakeStock{repo: pr}, nil, nil, nil, nil, nil, nil, nil, nil, nil)
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
	s := NewPartService(pr, &fakeFolderRepo{}, st, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	p, err := s.Create(context.Background(), CreatePartInput{SKU: "U1", Name: "Part", Category: "c", Unit: "шт", Price: "1"})
	if err != nil {
		t.Fatal(err)
	}
	sku := "U2"
	upd, err := s.Update(context.Background(), p.ID.String(), UpdatePartInput{SKU: &sku})
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
