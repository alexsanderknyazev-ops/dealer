package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/auth-service/internal/domain"
)

// UserRepository — хранилище пользователей в PostgreSQL.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository создаёт репозиторий пользователей.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create сохраняет пользователя.
func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, name, phone, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		u.ID, u.Email, u.PasswordHash, u.Name, u.Phone, u.Role, u.CreatedAt, u.UpdatedAt,
	)
	return err
}

// GetByEmail возвращает пользователя по email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, phone, role, created_at, updated_at
		FROM users WHERE email = $1
	`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByID возвращает пользователя по ID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, phone, role, created_at, updated_at
		FROM users WHERE id = $1
	`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
