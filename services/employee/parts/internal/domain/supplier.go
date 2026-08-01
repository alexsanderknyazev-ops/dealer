package domain

import (
	"time"

	"github.com/google/uuid"
)

type Supplier struct {
	ID        uuid.UUID
	Name      string
	INN       string
	Phone     string
	Email     string
	Notes     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
