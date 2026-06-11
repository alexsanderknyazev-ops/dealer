package domain

import (
	"time"

	"github.com/google/uuid"
)

const ClientRole = "client"

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FullName     string
	Phone        string
	Role         string // всегда client в JWT
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
