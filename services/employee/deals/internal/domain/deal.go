package domain

import (
	"time"

	"github.com/google/uuid"
)

type Deal struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	VehicleID  uuid.UUID
	Amount     string
	Stage      string // draft, in_progress, paid, completed, cancelled
	AssignedTo *uuid.UUID
	Notes      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
