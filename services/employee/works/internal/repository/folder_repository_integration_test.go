//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/works/internal/domain"
)

func TestFolderRepository_CRUD(t *testing.T) {
	repo := NewFolderRepository(testPool)
	ctx := context.Background()

	f := &domain.WorkFolder{
		ID:        uuid.New(),
		Name:      "IT Root Folder",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, f); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, f.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != f.Name || got.ParentID != nil {
		t.Fatalf("GetByID mismatch: %+v", got)
	}

	if _, err := repo.GetByID(ctx, uuid.New()); err != pgx.ErrNoRows {
		t.Fatalf("GetByID(missing): got %v want pgx.ErrNoRows", err)
	}

	parent := &domain.WorkFolder{
		ID:        uuid.New(),
		Name:      "IT Parent Folder",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, parent); err != nil {
		t.Fatalf("Create parent: %v", err)
	}

	f.ParentID = &parent.ID
	f.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, f); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err = repo.GetByID(ctx, f.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.ParentID == nil || *got.ParentID != parent.ID {
		t.Fatalf("Update parent not persisted: %+v", got)
	}

	if err := repo.Delete(ctx, parent.ID); err != nil {
		t.Fatalf("Delete parent: %v", err)
	}
	got, err = repo.GetByID(ctx, f.ID)
	if err != nil {
		t.Fatalf("GetByID after parent delete: %v", err)
	}
	if got.ParentID != nil {
		t.Fatalf("parent delete did not set null: %+v", got)
	}

	if err := repo.Delete(ctx, f.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, f.ID); err != pgx.ErrNoRows {
		t.Fatalf("GetByID after delete: got %v want pgx.ErrNoRows", err)
	}
}

func TestFolderRepository_ListByParent(t *testing.T) {
	repo := NewFolderRepository(testPool)
	ctx := context.Background()

	parent := &domain.WorkFolder{
		ID:        uuid.New(),
		Name:      "IT List Root Folder",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, parent); err != nil {
		t.Fatalf("Create parent: %v", err)
	}

	childA := &domain.WorkFolder{ID: uuid.New(), Name: "IT List Child A", ParentID: &parent.ID, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	childB := &domain.WorkFolder{ID: uuid.New(), Name: "IT List Child B", ParentID: &parent.ID, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	for _, c := range []*domain.WorkFolder{childA, childB} {
		if err := repo.Create(ctx, c); err != nil {
			t.Fatalf("Create child: %v", err)
		}
	}

	roots, err := repo.ListByParent(ctx, nil)
	if err != nil {
		t.Fatalf("ListByParent(nil): %v", err)
	}
	foundRoot := false
	for _, f := range roots {
		if f.ID == parent.ID {
			foundRoot = true
		}
		if f.ParentID != nil {
			t.Fatalf("ListByParent(nil) returned child %+v", f)
		}
	}
	if !foundRoot {
		t.Fatalf("ListByParent(nil) missing parent: %+v", roots)
	}

	children, err := repo.ListByParent(ctx, &parent.ID)
	if err != nil {
		t.Fatalf("ListByParent(parent): %v", err)
	}
	foundA, foundB := false, false
	for _, f := range children {
		if f.ID == childA.ID {
			foundA = true
		}
		if f.ID == childB.ID {
			foundB = true
		}
		if f.ParentID == nil || *f.ParentID != parent.ID {
			t.Fatalf("ListByParent(parent) returned wrong folder %+v", f)
		}
	}
	if !foundA || !foundB {
		t.Fatalf("ListByParent(parent) missing children: %+v", children)
	}
}
