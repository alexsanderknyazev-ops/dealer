package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/customers-service/internal/domain"
)

const testFmtGotErr = "got %v"

type fakeCustomerRepo struct {
	byID    map[uuid.UUID]*domain.Customer
	byEmail map[string]*domain.Customer
	list    []*domain.Customer
	total   int32
	err     error
	updErr  error
}

func (f *fakeCustomerRepo) Create(_ context.Context, c *domain.Customer) error {
	if f.err != nil {
		return f.err
	}
	if f.byID == nil {
		f.byID = make(map[uuid.UUID]*domain.Customer)
	}
	if f.byEmail == nil {
		f.byEmail = make(map[string]*domain.Customer)
	}
	cp := *c
	f.byID[c.ID] = &cp
	f.byEmail[c.Email] = &cp
	return nil
}

func (f *fakeCustomerRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Customer, error) {
	if f.err != nil {
		return nil, f.err
	}
	c, ok := f.byID[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return c, nil
}

func (f *fakeCustomerRepo) List(_ context.Context, _, _ int32, _ string) ([]*domain.Customer, int32, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.list, f.total, nil
}

func (f *fakeCustomerRepo) Update(_ context.Context, c *domain.Customer) error {
	if f.updErr != nil {
		return f.updErr
	}
	if f.err != nil {
		return f.err
	}
	if f.byID == nil {
		return errors.New("no db")
	}
	f.byID[c.ID] = c
	return nil
}

func (f *fakeCustomerRepo) Delete(_ context.Context, id uuid.UUID) error {
	if f.err != nil {
		return f.err
	}
	delete(f.byID, id)
	return nil
}

func TestCustomerService_Create_DefaultType(t *testing.T) {
	r := &fakeCustomerRepo{}
	s := NewCustomerService(r)
	ctx := context.Background()
	c, err := s.Create(ctx, CreateCustomerInput{Name: "A", Email: "a@b.c"})
	if err != nil {
		t.Fatal(err)
	}
	if c.CustomerType != "individual" {
		t.Fatalf("want individual, got %q", c.CustomerType)
	}
}

func TestCustomerService_Get_ParseErr(t *testing.T) {
	s := NewCustomerService(&fakeCustomerRepo{})
	_, err := s.Get(context.Background(), "not-a-uuid")
	if err != ErrNotFound {
		t.Fatalf(testFmtGotErr, err)
	}
}

func TestCustomerService_Get_NotFound(t *testing.T) {
	s := NewCustomerService(&fakeCustomerRepo{})
	_, err := s.Get(context.Background(), uuid.New().String())
	if err != ErrNotFound {
		t.Fatalf(testFmtGotErr, err)
	}
}

func TestCustomerService_Get_OK(t *testing.T) {
	id := uuid.New()
	now := time.Now().UTC()
	c := &domain.Customer{ID: id, Name: "N", Email: "e@e.e", CreatedAt: now, UpdatedAt: now}
	r := &fakeCustomerRepo{byID: map[uuid.UUID]*domain.Customer{id: c}}
	s := NewCustomerService(r)
	got, err := s.Get(context.Background(), id.String())
	if err != nil || got.ID != id {
		t.Fatalf("err=%v id=%v", err, got)
	}
}

func TestCustomerService_List_LimitClamp(t *testing.T) {
	r := &fakeCustomerRepo{list: []*domain.Customer{}, total: 0}
	s := NewCustomerService(r)
	_, _, err := s.List(context.Background(), 0, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = s.List(context.Background(), 200, 0, "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCustomerService_Update_NotFound(t *testing.T) {
	s := NewCustomerService(&fakeCustomerRepo{})
	n := "x"
	_, err := s.Update(context.Background(), uuid.New().String(), UpdateCustomerInput{Name: &n})
	if err != ErrNotFound {
		t.Fatalf(testFmtGotErr, err)
	}
}

func TestCustomerService_Update_OK(t *testing.T) {
	id := uuid.New()
	now := time.Now().UTC()
	c := &domain.Customer{ID: id, Name: "Old", Email: "e@e.e", CreatedAt: now, UpdatedAt: now}
	r := &fakeCustomerRepo{byID: map[uuid.UUID]*domain.Customer{id: c}}
	s := NewCustomerService(r)
	n := "New"
	got, err := s.Update(context.Background(), id.String(), UpdateCustomerInput{Name: &n})
	if err != nil || got.Name != "New" {
		t.Fatalf("err=%v name=%q", err, got.Name)
	}
}

func TestCustomerService_Delete_ParseErr(t *testing.T) {
	s := NewCustomerService(&fakeCustomerRepo{})
	if err := s.Delete(context.Background(), "bad"); err != ErrNotFound {
		t.Fatalf(testFmtGotErr, err)
	}
}

func TestCustomerService_Create_RepoErr(t *testing.T) {
	r := &fakeCustomerRepo{err: errors.New("db")}
	s := NewCustomerService(r)
	_, err := s.Create(context.Background(), CreateCustomerInput{Name: "A", Email: "a@b.c"})
	if err == nil {
		t.Fatal("want err")
	}
}

func TestCustomerService_Get_RepoErr(t *testing.T) {
	r := &fakeCustomerRepo{err: errors.New("db down")}
	s := NewCustomerService(r)
	_, err := s.Get(context.Background(), uuid.New().String())
	if err == nil || err == ErrNotFound {
		t.Fatalf(testFmtGotErr, err)
	}
}

func TestCustomerService_List_Err(t *testing.T) {
	r := &fakeCustomerRepo{err: errors.New("db")}
	s := NewCustomerService(r)
	_, _, err := s.List(context.Background(), 10, 0, "")
	if err == nil {
		t.Fatal("want err")
	}
}

func TestCustomerService_Update_RepoErrOnSave(t *testing.T) {
	id := uuid.New()
	now := time.Now().UTC()
	c := &domain.Customer{ID: id, Name: "O", Email: "e@e", CreatedAt: now, UpdatedAt: now}
	r := &fakeCustomerRepo{byID: map[uuid.UUID]*domain.Customer{id: c}, updErr: errors.New("write fail")}
	s := NewCustomerService(r)
	n := "x"
	_, err := s.Update(context.Background(), id.String(), UpdateCustomerInput{Name: &n})
	if err == nil || err == ErrNotFound {
		t.Fatalf(testFmtGotErr, err)
	}
}
