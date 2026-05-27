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

// CreateCustomerInput is the payload for Create (keeps CustomerAPI arity within Sonar limits).
type CreateCustomerInput struct {
	Name         string
	Email        string
	Phone        string
	CustomerType string
	INN          string
	Address      string
	Notes        string
}

// UpdateCustomerInput holds optional fields for Update.
type UpdateCustomerInput struct {
	Name         *string
	Email        *string
	Phone        *string
	CustomerType *string
	INN          *string
	Address      *string
	Notes        *string
}

// CustomerAPI — контракт для HTTP/gRPC (моки в тестах).
type CustomerAPI interface {
	Create(ctx context.Context, in CreateCustomerInput) (*domain.Customer, error)
	Get(ctx context.Context, id string) (*domain.Customer, error)
	List(ctx context.Context, p domain.CustomerListParams) ([]*domain.Customer, int32, error)
	Update(ctx context.Context, id string, in UpdateCustomerInput) (*domain.Customer, error)
	Delete(ctx context.Context, id string) error
}

type customerRepository interface {
	Create(ctx context.Context, c *domain.Customer) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error)
	List(ctx context.Context, p domain.CustomerListParams) ([]*domain.Customer, int32, error)
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

func (s *CustomerService) Create(ctx context.Context, in CreateCustomerInput) (*domain.Customer, error) {
	customerType := in.CustomerType
	if customerType == "" {
		customerType = "individual"
	}
	now := time.Now().UTC()
	c := &domain.Customer{
		ID:           uuid.New(),
		Name:         in.Name,
		Email:        in.Email,
		Phone:        in.Phone,
		CustomerType: customerType,
		INN:          in.INN,
		Address:      in.Address,
		Notes:        in.Notes,
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

func (s *CustomerService) List(ctx context.Context, p domain.CustomerListParams) ([]*domain.Customer, int32, error) {
	lp := p
	if lp.Limit <= 0 || lp.Limit > 100 {
		lp.Limit = 20
	}
	return s.repo.List(ctx, lp)
}

func (s *CustomerService) Update(ctx context.Context, id string, in UpdateCustomerInput) (*domain.Customer, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	existing, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, ErrNotFound
	}
	if in.Name != nil {
		existing.Name = *in.Name
	}
	if in.Email != nil {
		existing.Email = *in.Email
	}
	if in.Phone != nil {
		existing.Phone = *in.Phone
	}
	if in.CustomerType != nil {
		existing.CustomerType = *in.CustomerType
	}
	if in.INN != nil {
		existing.INN = *in.INN
	}
	if in.Address != nil {
		existing.Address = *in.Address
	}
	if in.Notes != nil {
		existing.Notes = *in.Notes
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
