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

	"github.com/dealer/dealer/auth-service/internal/domain"
)

func TestUserRepository_CRUD(t *testing.T) {
	repo := NewUserRepository(testPool)
	ctx := context.Background()

	u := &domain.User{
		ID:           uuid.New(),
		Email:        "it.user@example.com",
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuv",
		Name:         "IT User",
		Phone:        "+79990001122",
		Role:         domain.DefaultRole,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Email != u.Email || got.Name != u.Name || got.Phone != u.Phone || got.Role != u.Role {
		t.Fatalf("GetByID mismatch: got %+v want %+v", got, u)
	}

	byEmail, err := repo.GetByEmail(ctx, u.Email)
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if byEmail.ID != u.ID || byEmail.PasswordHash != u.PasswordHash {
		t.Fatalf("GetByEmail mismatch: %+v", byEmail)
	}

	if _, err := repo.GetByID(ctx, uuid.New()); err != pgx.ErrNoRows {
		t.Fatalf("GetByID(missing): got %v want pgx.ErrNoRows", err)
	}
}

func TestUserRepository_UniqueEmail(t *testing.T) {
	repo := NewUserRepository(testPool)
	ctx := context.Background()

	u := &domain.User{
		ID:           uuid.New(),
		Email:        "it.unique@example.com",
		PasswordHash: "hash",
		Name:         "Unique",
		Role:         domain.DefaultRole,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	dup := *u
	dup.ID = uuid.New()
	err := repo.Create(ctx, &dup)
	if err == nil {
		t.Fatal("expected unique violation for duplicate email")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("expected SQLSTATE 23505, got %v", err)
	}
}
