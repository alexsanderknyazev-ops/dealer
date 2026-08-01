//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/dealer/dealer/services/works/internal/domain"
)

func TestWorkRepository_CRUD(t *testing.T) {
	repo := NewWorkRepository(testPool)
	ctx := context.Background()

	w := &domain.Work{
		ID:         uuid.New(),
		Code:       "ITW-001",
		Name:       "IT Work",
		Category:   "ТО",
		LaborHours: "2.5",
		UnitPrice:  "1500.00",
		Notes:      "integration test",
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := repo.Create(ctx, w); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, w.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Code != w.Code || got.Name != w.Name || got.Category != w.Category || got.LaborHours != "2.500" || got.UnitPrice != "1500.00" {
		t.Fatalf("GetByID mismatch: %+v", got)
	}

	if _, err := repo.GetByID(ctx, uuid.New()); err != pgx.ErrNoRows {
		t.Fatalf("GetByID(missing): got %v want pgx.ErrNoRows", err)
	}

	w.Code = "ITW-001-R"
	w.Name = "IT Work Renamed"
	w.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, w); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err = repo.GetByID(ctx, w.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Code != "ITW-001-R" || got.Name != "IT Work Renamed" {
		t.Fatalf("Update not persisted: %+v", got)
	}

	if err := repo.Delete(ctx, w.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, w.ID); err != pgx.ErrNoRows {
		t.Fatalf("GetByID after delete: got %v want pgx.ErrNoRows", err)
	}
}

func TestWorkRepository_ListFilters(t *testing.T) {
	folderRepo := NewFolderRepository(testPool)
	ctx := context.Background()

	folder := &domain.WorkFolder{
		ID:        uuid.New(),
		Name:      "IT Work Folder",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := folderRepo.Create(ctx, folder); err != nil {
		t.Fatalf("Create folder: %v", err)
	}

	repo := NewWorkRepository(testPool)
	works := []*domain.Work{
		{ID: uuid.New(), Code: "ITW-FLT1", Name: "Filter Work One", Category: "ТО", FolderID: &folder.ID, LaborHours: "1.0", UnitPrice: "1000.00", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		{ID: uuid.New(), Code: "ITW-FLT2", Name: "Filter Work Two", Category: "Тормоза", LaborHours: "2.0", UnitPrice: "2000.00", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		{ID: uuid.New(), Code: "ITW-FLT3", Name: "Brake Work", Category: "Тормоза", FolderID: &folder.ID, LaborHours: "3.0", UnitPrice: "3000.00", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
	}
	for _, w := range works {
		if err := repo.Create(ctx, w); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	list, total, err := repo.List(ctx, 10, 0, "Filter Work", "", nil)
	if err != nil {
		t.Fatalf("List search: %v", err)
	}
	if total < 2 || len(list) < 2 {
		t.Fatalf("List search: total=%d len=%d want >=2", total, len(list))
	}

	list, total, err = repo.List(ctx, 10, 0, "", "Тормоза", nil)
	if err != nil {
		t.Fatalf("List category: %v", err)
	}
	if total < 2 {
		t.Fatalf("List category total: got %d want >=2", total)
	}
	for _, w := range list {
		if w.Category != "Тормоза" {
			t.Fatalf("List category filter returned %q", w.Category)
		}
	}

	list, total, err = repo.List(ctx, 10, 0, "", "", &folder.ID)
	if err != nil {
		t.Fatalf("List folder: %v", err)
	}
	if total < 2 {
		t.Fatalf("List folder total: got %d want >=2", total)
	}
	for _, w := range list {
		if w.FolderID == nil || *w.FolderID != folder.ID {
			t.Fatalf("List folder filter returned %+v", w)
		}
	}

	page, total, err := repo.List(ctx, 2, 0, "", "", nil)
	if err != nil {
		t.Fatalf("List paged: %v", err)
	}
	if len(page) > 2 {
		t.Fatalf("List page len: got %d want <=2", len(page))
	}
	if total < 3 {
		t.Fatalf("List page total: got %d want >=3", total)
	}
}

func TestWorkRepository_UniqueCode(t *testing.T) {
	repo := NewWorkRepository(testPool)
	ctx := context.Background()

	w := &domain.Work{
		ID:         uuid.New(),
		Code:       "ITW-UNIQUE",
		Name:       "Unique Work",
		Category:   "ТО",
		LaborHours: "1.0",
		UnitPrice:  "100.00",
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := repo.Create(ctx, w); err != nil {
		t.Fatalf("Create: %v", err)
	}

	dup := *w
	dup.ID = uuid.New()
	err := repo.Create(ctx, &dup)
	if err == nil {
		t.Fatal("expected unique violation for duplicate code")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("expected SQLSTATE 23505, got %v", err)
	}
}
