package domain

import (
	"time"

	"github.com/google/uuid"
)

type DealerPoint struct {
	ID        uuid.UUID
	Name      string
	Address   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
