//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/customers-service/internal/domain"
)

func TestCustomerRepository_CRUD(t *testing.T) {
	repo := NewCustomerRepository(testPool)
	ctx := context.Background()

	c := &domain.Customer{
		ID:           uuid.New(),
		Name:         "IT Customer",
		Email:        "it.customer@example.com",
		Phone:        "+79990000001",
		CustomerType: "individual",
		INN:          "7700000000",
		Address:      "Moscow",
		Notes:        "integration test",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != c.Name || got.Email != c.Email || got.CustomerType != c.CustomerType || got.INN != c.INN {
		t.Fatalf("GetByID mismatch: %+v", got)
	}

	if _, err := repo.GetByID(ctx, uuid.New()); err != pgx.ErrNoRows {
		t.Fatalf("GetByID(missing): got %v want pgx.ErrNoRows", err)
	}
}

func TestCustomerRepository_ListSearchAndPagination(t *testing.T) {
	repo := NewCustomerRepository(testPool)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		c := &domain.Customer{
			ID:           uuid.New(),
			Name:         "Searchable Client " + string(rune('A'+i)),
			Email:        "search" + string(rune('a'+i)) + "@example.com",
			Phone:        "+7999000000" + string(rune('1'+i)),
			CustomerType: "individual",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}
		if err := repo.Create(ctx, c); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	list, total, err := repo.List(ctx, domain.CustomerListParams{Limit: 10, Offset: 0, Search: "Searchable Client"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total < 3 {
		t.Fatalf("List total: got %d want >=3", total)
	}
	if len(list) != 3 {
		t.Fatalf("List len: got %d want 3", len(list))
	}

	page, total, err := repo.List(ctx, domain.CustomerListParams{Limit: 2, Offset: 0})
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

func TestCustomerRepository_UpdateAndDelete(t *testing.T) {
	repo := NewCustomerRepository(testPool)
	ctx := context.Background()

	c := &domain.Customer{
		ID:           uuid.New(),
		Name:         "Before",
		Email:        "before@example.com",
		CustomerType: "individual",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create: %v", err)
	}

	c.Name = "After"
	c.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, c); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err := repo.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Name != "After" {
		t.Fatalf("Update not persisted: %+v", got)
	}

	if err := repo.Delete(ctx, c.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, c.ID); err != pgx.ErrNoRows {
		t.Fatalf("GetByID after delete: got %v want pgx.ErrNoRows", err)
	}
}
