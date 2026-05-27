package domain

import (
	"time"

	"github.com/google/uuid"
)

type Warehouse struct {
	ID            uuid.UUID
	DealerPointID uuid.UUID
	LegalEntityID uuid.UUID
	Type          string // "cars" | "parts"
	Name          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
