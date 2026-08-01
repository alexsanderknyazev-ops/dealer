//go:build integration

package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/dealer/dealer/services/employees/internal/domain"
)

func TestEmployeeRepository_CRUD(t *testing.T) {
	repo := NewEmployeeRepository(testPool)
	ctx := context.Background()

	userID := uuid.New()
	e := &domain.Employee{
		ID:         uuid.New(),
		UserID:     &userID,
		FullName:   "IT Employee",
		Position:   "master",
		Department: "СТО",
		Phone:      "+79990001111",
		Active:     true,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := repo.Create(ctx, e); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, e.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.FullName != e.FullName || got.Position != e.Position || got.Department != e.Department || got.Phone != e.Phone || got.Active != true {
		t.Fatalf("GetByID mismatch: %+v", got)
	}
	if got.UserID == nil || *got.UserID != userID {
		t.Fatalf("GetByID user_id mismatch: %+v", got)
	}

	if _, err := repo.GetByID(ctx, uuid.New()); err != pgx.ErrNoRows {
		t.Fatalf("GetByID(missing): got %v want pgx.ErrNoRows", err)
	}

	e.FullName = "IT Employee Renamed"
	e.Position = "sales"
	e.Active = false
	e.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, e); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err = repo.GetByID(ctx, e.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.FullName != "IT Employee Renamed" || got.Position != "sales" || got.Active != false {
		t.Fatalf("Update not persisted: %+v", got)
	}

	if err := repo.Delete(ctx, e.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, e.ID); err != pgx.ErrNoRows {
		t.Fatalf("GetByID after delete: got %v want pgx.ErrNoRows", err)
	}
}

func TestEmployeeRepository_GetByUserIDAndResolveRef(t *testing.T) {
	repo := NewEmployeeRepository(testPool)
	ctx := context.Background()

	userID := uuid.New()
	e := &domain.Employee{
		ID:         uuid.New(),
		UserID:     &userID,
		FullName:   "IT Resolve Employee",
		Position:   "sales",
		Department: "Продажи",
		Phone:      "+79990001112",
		Active:     true,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := repo.Create(ctx, e); err != nil {
		t.Fatalf("Create: %v", err)
	}

	byUser, err := repo.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if byUser.ID != e.ID {
		t.Fatalf("GetByUserID mismatch: %+v", byUser)
	}

	if _, err := repo.GetByUserID(ctx, uuid.New()); err != pgx.ErrNoRows {
		t.Fatalf("GetByUserID(missing): got %v want pgx.ErrNoRows", err)
	}

	byID, err := repo.ResolveRef(ctx, e.ID)
	if err != nil {
		t.Fatalf("ResolveRef(id): %v", err)
	}
	if byID.ID != e.ID {
		t.Fatalf("ResolveRef(id) mismatch: %+v", byID)
	}

	byUser, err = repo.ResolveRef(ctx, userID)
	if err != nil {
		t.Fatalf("ResolveRef(userID): %v", err)
	}
	if byUser.ID != e.ID {
		t.Fatalf("ResolveRef(userID) mismatch: %+v", byUser)
	}

	if _, err := repo.ResolveRef(ctx, uuid.New()); err != pgx.ErrNoRows {
		t.Fatalf("ResolveRef(missing): got %v want pgx.ErrNoRows", err)
	}
}

func TestEmployeeRepository_ListFilters(t *testing.T) {
	repo := NewEmployeeRepository(testPool)
	ctx := context.Background()

	employees := []*domain.Employee{
		{ID: uuid.New(), FullName: "IT List One", Position: "master", Department: "СТО", Phone: "+79990001121", Active: true, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		{ID: uuid.New(), FullName: "IT List Two", Position: "master", Department: "Продажи", Phone: "+79990001122", Active: false, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		{ID: uuid.New(), FullName: "IT List Three", Position: "sales", Department: "СТО", Phone: "+79990001123", Active: true, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
	}
	for _, e := range employees {
		if err := repo.Create(ctx, e); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	list, total, err := repo.List(ctx, 10, 0, "IT List", "", false)
	if err != nil {
		t.Fatalf("List search: %v", err)
	}
	if total < 3 || len(list) < 3 {
		t.Fatalf("List search: total=%d len=%d want >=3", total, len(list))
	}

	list, total, err = repo.List(ctx, 10, 0, "", "master", false)
	if err != nil {
		t.Fatalf("List position: %v", err)
	}
	if total < 2 {
		t.Fatalf("List position total: got %d want >=2", total)
	}
	for _, e := range list {
		if e.Position != "master" {
			t.Fatalf("List position filter returned %q", e.Position)
		}
	}

	list, total, err = repo.List(ctx, 10, 0, "", "", true)
	if err != nil {
		t.Fatalf("List active only: %v", err)
	}
	if total < 2 {
		t.Fatalf("List active only total: got %d want >=2", total)
	}
	for _, e := range list {
		if !e.Active {
			t.Fatalf("List active only returned inactive employee %+v", e)
		}
	}

	page, total, err := repo.List(ctx, 2, 0, "", "", false)
	if err != nil {
		t.Fatalf("List paged: %v", err)
	}
	if len(page) > 2 {
		t.Fatalf("List page len: got %d want <=2", len(page))
	}
	if total < 3 {
		t.Fatalf("List page total: got %d want >=3", total)
	}

	list, total, err = repo.List(ctx, 10, 0, fmt.Sprintf("IT List %d", 42), "", false)
	if err != nil {
		t.Fatalf("List search miss: %v", err)
	}
	if total != 0 {
		t.Fatalf("List search miss total: got %d want 0", total)
	}
}

func TestEmployeeRepository_UniqueUserID(t *testing.T) {
	repo := NewEmployeeRepository(testPool)
	ctx := context.Background()

	userID := uuid.New()
	e := &domain.Employee{
		ID:         uuid.New(),
		UserID:     &userID,
		FullName:   "IT Unique User",
		Position:   "master",
		Department: "СТО",
		Phone:      "+79990001131",
		Active:     true,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := repo.Create(ctx, e); err != nil {
		t.Fatalf("Create: %v", err)
	}

	dup := *e
	dup.ID = uuid.New()
	err := repo.Create(ctx, &dup)
	if err == nil {
		t.Fatal("expected unique violation for duplicate user_id")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("expected SQLSTATE 23505, got %v", err)
	}
}
