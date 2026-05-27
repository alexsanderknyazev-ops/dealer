package domain

import (
	"time"

	"github.com/google/uuid"
)

// User — сущность пользователя.
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Name         string
	Phone        string
	Role         string // admin, manager, sales, accountant, viewer, warranty_engineer, parts_manager, storekeeper, master, consultant, cashier
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// DefaultRole — роль по умолчанию при регистрации.
const DefaultRole = "sales"
