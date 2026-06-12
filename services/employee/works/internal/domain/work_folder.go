package domain

import (
	"time"

	"github.com/google/uuid"
)

type WorkFolder struct {
	ID        uuid.UUID
	Name      string
	ParentID  *uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}
