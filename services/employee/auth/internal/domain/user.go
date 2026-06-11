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
	Role         string // staff-роли или client
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// DefaultRole — роль по умолчанию при регистрации сотрудника.
const DefaultRole = "sales"

// ClientRole — роль владельца авто (самостоятельная регистрация).
const ClientRole = "client"
