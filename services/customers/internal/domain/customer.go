package domain

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID           uuid.UUID
	Name         string
	Email        string
	Phone        string
	CustomerType string // individual, legal
	INN          string
	Address      string
	Notes        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CustomerListParams is pagination and search for listing customers.
type CustomerListParams struct {
	Limit, Offset int32
	Search        string
}
