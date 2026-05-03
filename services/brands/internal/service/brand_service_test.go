package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/brands/internal/domain"
)

type fakeBrandRepo struct {
	byID   map[uuid.UUID]*domain.Brand
	list   []*domain.Brand
	total  int32
	err    error
	updErr error
}

func (f *fakeBrandRepo) Create(_ context.Context, b *domain.Brand) error {
	if f.err != nil {
		return f.err
	}
	if f.byID == nil {
		f.byID = make(map[uuid.UUID]*domain.Brand)
	}
	cp := *b
	f.byID[b.ID] = &cp
	return nil
}

func (f *fakeBrandRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Brand, error) {
	if f.err != nil {
		return nil, f.err
	}
	b, ok := f.byID[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return b, nil
}

func (f *fakeBrandRepo) List(_ context.Context, _, _ int32, _ string) ([]*domain.Brand, int32, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.list, f.total, nil
}

func (f *fakeBrandRepo) Update(_ context.Context, b *domain.Brand) error {
	if f.updErr != nil {
		return f.updErr
	}
	if f.byID == nil {
		return errors.New("no db")
	}
	f.byID[b.ID] = b
	return nil
}

func (f *fakeBrandRepo) Delete(_ context.Context, id uuid.UUID) error {
	if f.err != nil {
		return f.err
	}
	delete(f.byID, id)
	return nil
}

func TestBrandService_CRUD(t *testing.T) {
	ctx := context.Background()
	r := &fakeBrandRepo{byID: make(map[uuid.UUID]*domain.Brand)}
	s := NewBrandService(r)
	b, err := s.Create(ctx, "Toyota")
	if err != nil || b.Name != "Toyota" {
		t.Fatal(err, b)
	}
	got, err := s.Get(ctx, b.ID.String())
	if err != nil || got.Name != "Toyota" {
		t.Fatal(err)
	}
	_, _, err = s.List(ctx, 0, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	n := "Lexus"
	_, err = s.Update(ctx, b.ID.String(), &n)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, b.ID.String()); err != nil {
		t.Fatal(err)
	}
}

func TestBrandService_Get_ParseErr(t *testing.T) {
	_, err := NewBrandService(&fakeBrandRepo{}).Get(context.Background(), "x")
	if err != ErrNotFound {
		t.Fatalf("%v", err)
	}
}

func TestBrandService_Get_DBErr(t *testing.T) {
	_, err := NewBrandService(&fakeBrandRepo{err: errors.New("db")}).Get(context.Background(), uuid.New().String())
	if err == nil || err == ErrNotFound {
		t.Fatal(err)
	}
}

func TestBrandService_Update_NotFound(t *testing.T) {
	_, err := NewBrandService(&fakeBrandRepo{}).Update(context.Background(), "bad-uuid", ptr("n"))
	if err != ErrNotFound {
		t.Fatal(err)
	}
}

func TestBrandService_Update_SaveErr(t *testing.T) {
	id := uuid.New()
	now := time.Now().UTC()
	b := &domain.Brand{ID: id, Name: "a", CreatedAt: now, UpdatedAt: now}
	r := &fakeBrandRepo{byID: map[uuid.UUID]*domain.Brand{id: b}, updErr: errors.New("w")}
	n := "b"
	_, err := NewBrandService(r).Update(context.Background(), id.String(), &n)
	if err == nil {
		t.Fatal("want err")
	}
}

func ptr(s string) *string { return &s }
