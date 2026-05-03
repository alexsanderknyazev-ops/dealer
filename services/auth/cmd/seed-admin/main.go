// Команда создаёт пользователя admin, если такого ещё нет.
//
// Использование (POSTGRES_DSN обязателен, без значения по умолчанию):
//
//	POSTGRES_DSN="postgres://USER:PASSWORD@HOST:PORT/dealer?sslmode=disable" \
//	ADMIN_EMAIL=admin@dealer.local \
//	ADMIN_PASSWORD=admin123 \
//	go run ./cmd/seed-admin
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/dealer/dealer/auth-service/internal/domain"
	"github.com/dealer/dealer/auth-service/internal/repository"
	"github.com/dealer/dealer/pkg/postgres"
)

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Fatal("POSTGRES_DSN is required (no default; see .env.example)")
	}
	email := os.Getenv("ADMIN_EMAIL")
	if email == "" {
		email = "admin@dealer.local"
	}
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "admin123"
	}

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, dsn)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	repo := repository.NewUserRepository(pool)
	existing, err := repo.GetByEmail(ctx, email)
	if err == nil {
		log.Printf("Пользователь %s уже существует (id=%s, role=%s). Ничего не делаем.", email, existing.ID, existing.Role)
		return
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Fatalf("get user: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("bcrypt: %v", err)
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         "Администратор",
		Phone:        "",
		Role:         "admin",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := repo.Create(ctx, user); err != nil {
		log.Fatalf("create user: %v", err)
	}
	fmt.Printf("Создан пользователь admin: %s (роль: admin). Смените пароль после первого входа.\n", email)
}
