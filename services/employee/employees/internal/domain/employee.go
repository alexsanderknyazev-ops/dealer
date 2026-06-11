package domain

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID         uuid.UUID
	UserID     *uuid.UUID
	FullName   string
	Position   string
	Department string
	Phone      string
	Active     bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
