package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/customers-service/internal/domain"
)

var ErrNotFound = errors.New("customer not found")

// CustomerAPI — контракт для HTTP/gRPC (моки в тестах).
type CustomerAPI interface {
	Create(ctx context.Context, name, email, phone, customerType, inn, address, notes string) (*domain.Customer, error)
	Get(ctx context.Context, id string) (*domain.Customer, error)
	List(ctx context.Context, limit, offset int32, search string) ([]*domain.Customer, int32, error)
	Update(ctx context.Context, id string, name, email, phone, customerType, inn, address, notes *string) (*domain.Customer, error)
	Delete(ctx context.Context, id string) error
}

type customerRepository interface {
	Create(ctx context.Context, c *domain.Customer) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error)
	List(ctx context.Context, limit, offset int32, search string) ([]*domain.Customer, int32, error)
	Update(ctx context.Context, c *domain.Customer) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type CustomerService struct {
	repo customerRepository
}

func NewCustomerService(repo customerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

var _ CustomerAPI = (*CustomerService)(nil)

func (s *CustomerService) Create(ctx context.Context, name, email, phone, customerType, inn, address, notes string) (*domain.Customer, error) {
	if customerType == "" {
		customerType = "individual"
	}
	now := time.Now().UTC()
	c := &domain.Customer{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		Phone:        phone,
		CustomerType: customerType,
		INN:          inn,
		Address:      address,
		Notes:        notes,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CustomerService) Get(ctx context.Context, id string) (*domain.Customer, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	c, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return c, nil
}

func (s *CustomerService) List(ctx context.Context, limit, offset int32, search string) ([]*domain.Customer, int32, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, limit, offset, search)
}

func (s *CustomerService) Update(ctx context.Context, id string, name, email, phone, customerType, inn, address, notes *string) (*domain.Customer, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	existing, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, ErrNotFound
	}
	if name != nil {
		existing.Name = *name
	}
	if email != nil {
		existing.Email = *email
	}
	if phone != nil {
		existing.Phone = *phone
	}
	if customerType != nil {
		existing.CustomerType = *customerType
	}
	if inn != nil {
		existing.INN = *inn
	}
	if address != nil {
		existing.Address = *address
	}
	if notes != nil {
		existing.Notes = *notes
	}
	existing.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *CustomerService) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrNotFound
	}
	return s.repo.Delete(ctx, uid)
}
