//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/client-auth-service/internal/domain"
)

func TestUserRepository_Roundtrip(t *testing.T) {
	repo := NewUserRepository(testPool)
	ctx := context.Background()

	u := &domain.User{
		ID:           uuid.New(),
		Email:        "it.client.auth.roundtrip@example.com",
		PasswordHash: "$2a$10$roundtriphashexample1234567890",
		FullName:     "IT Auth User",
		Phone:        "+79990000001",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	byEmail, err := repo.GetByEmail(ctx, u.Email)
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if byEmail.ID != u.ID || byEmail.Email != u.Email || byEmail.PasswordHash != u.PasswordHash ||
		byEmail.FullName != u.FullName || byEmail.Phone != u.Phone {
		t.Fatalf("GetByEmail mismatch: got %+v want %+v", byEmail, u)
	}
	if byEmail.Role != domain.ClientRole {
		t.Fatalf("Role: got %q want %q", byEmail.Role, domain.ClientRole)
	}

	byID, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if byID.ID != u.ID || byID.Email != u.Email || byID.FullName != u.FullName {
		t.Fatalf("GetByID mismatch: got %+v want %+v", byID, u)
	}

	if _, err := repo.GetByID(ctx, uuid.New()); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("GetByID missing: got %v want pgx.ErrNoRows", err)
	}
	if _, err := repo.GetByEmail(ctx, "it.client.auth.missing@example.com"); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("GetByEmail missing: got %v want pgx.ErrNoRows", err)
	}
}

func TestUserRepository_DuplicateEmail(t *testing.T) {
	repo := NewUserRepository(testPool)
	ctx := context.Background()

	u := &domain.User{
		ID:           uuid.New(),
		Email:        "it.client.auth.unique@example.com",
		PasswordHash: "hash",
		FullName:     "Unique",
		Phone:        "+79990000002",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	dup := *u
	dup.ID = uuid.New()
	if err := repo.Create(ctx, &dup); !errors.Is(err, ErrEmailExists) {
		t.Fatalf("duplicate Create: got %v want ErrEmailExists", err)
	}
}

func TestUserRepository_EnsureExistsIdempotent(t *testing.T) {
	repo := NewUserRepository(testPool)
	ctx := context.Background()

	originalEmail := "it.client.auth.ensure@example.com"
	u := &domain.User{
		ID:           uuid.New(),
		Email:        originalEmail,
		PasswordHash: "hash",
		FullName:     "Ensure",
		Phone:        "+79990000003",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := repo.EnsureExists(ctx, u); err != nil {
		t.Fatalf("first EnsureExists: %v", err)
	}

	u.Email = "it.client.auth.ensure.changed@example.com"
	if err := repo.EnsureExists(ctx, u); err != nil {
		t.Fatalf("second EnsureExists: %v", err)
	}

	var count int
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE id = $1`, u.ID).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("rows after double EnsureExists: got %d want 1", count)
	}

	got, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Email != originalEmail {
		t.Fatalf("EnsureExists must not overwrite existing row: got email %q want %q", got.Email, originalEmail)
	}
}
