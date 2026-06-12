package domain

import (
	"time"

	"github.com/google/uuid"
)

type ClientNotification struct {
	ID         uuid.UUID
	ClientID   uuid.UUID
	UserID     uuid.UUID
	Kind       string
	SourceType string
	SourceID   uuid.UUID
	Title      string
	Body       string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
