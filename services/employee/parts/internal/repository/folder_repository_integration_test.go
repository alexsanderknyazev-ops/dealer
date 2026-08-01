//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/parts/internal/domain"
)

func TestFolderRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewFolderRepository(testPool)

	now := time.Now().UTC()
	parent := &domain.PartFolder{ID: uuid.New(), Name: "Родитель", CreatedAt: now, UpdatedAt: now}
	child := &domain.PartFolder{ID: uuid.New(), Name: "Дочерняя", ParentID: &parent.ID, CreatedAt: now, UpdatedAt: now}
	if err := repo.Create(ctx, parent); err != nil {
		t.Fatalf("Create parent: %v", err)
	}
	if err := repo.Create(ctx, child); err != nil {
		t.Fatalf("Create child: %v", err)
	}

	got, err := repo.GetByID(ctx, child.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Дочерняя" || got.ParentID == nil || *got.ParentID != parent.ID {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}

	roots, err := repo.ListByParent(ctx, nil)
	if err != nil {
		t.Fatalf("ListByParent nil: %v", err)
	}
	foundRoot := false
	for _, f := range roots {
		if f.ID == parent.ID {
			foundRoot = true
		}
	}
	if !foundRoot {
		t.Fatal("parent should be a root folder")
	}
	children, err := repo.ListByParent(ctx, &parent.ID)
	if err != nil {
		t.Fatalf("ListByParent parent: %v", err)
	}
	if len(children) != 1 || children[0].ID != child.ID {
		t.Fatalf("children mismatch: %+v", children)
	}

	child.Name = "переименована"
	child.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, child); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ = repo.GetByID(ctx, child.ID)
	if got.Name != "переименована" {
		t.Fatalf("update not persisted: %+v", got)
	}

	if err := repo.Delete(ctx, parent.ID); err != nil {
		t.Fatalf("Delete parent: %v", err)
	}
	if _, err := repo.GetByID(ctx, parent.ID); err != pgx.ErrNoRows {
		t.Fatalf("expected ErrNoRows after delete, got %v", err)
	}
	childAfter, err := repo.GetByID(ctx, child.ID)
	if err != nil {
		t.Fatalf("GetByID child: %v", err)
	}
	if childAfter.ParentID != nil {
		t.Fatal("child parent_id should be SET NULL after parent delete")
	}
}
