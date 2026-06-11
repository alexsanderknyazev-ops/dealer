package domain

import (
	"time"

	"github.com/google/uuid"
)

type Work struct {
	ID         uuid.UUID
	Code       string
	Name       string
	Category   string
	LaborHours string
	UnitPrice  string
	Notes      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
