package domain

import (
	"time"

	"github.com/google/uuid"
)

type LegalEntity struct {
	ID        uuid.UUID
	Name      string
	INN       string
	Address   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
