package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/employees/internal/domain"
)

var ErrNotFound = errors.New("employee not found")

type EmployeeAPI interface {
	Create(ctx context.Context, userID, fullName, position, department, phone string, active bool) (*domain.Employee, error)
	Get(ctx context.Context, id string) (*domain.Employee, error)
	GetByUserID(ctx context.Context, userID string) (*domain.Employee, error)
	ResolveRef(ctx context.Context, ref string) (*domain.Employee, error)
	List(ctx context.Context, limit, offset int32, search, position string, activeOnly bool) ([]*domain.Employee, int32, error)
	Update(ctx context.Context, id string, userID, fullName, position, department, phone *string, active *bool) (*domain.Employee, error)
	Delete(ctx context.Context, id string) error
}

type employeeRepository interface {
	Create(ctx context.Context, e *domain.Employee) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Employee, error)
	ResolveRef(ctx context.Context, ref uuid.UUID) (*domain.Employee, error)
	List(ctx context.Context, limit, offset int32, search, position string, activeOnly bool) ([]*domain.Employee, int32, error)
	Update(ctx context.Context, e *domain.Employee) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type EmployeeService struct {
	repo employeeRepository
}

func NewEmployeeService(repo employeeRepository) *EmployeeService {
	return &EmployeeService{repo: repo}
}

var _ EmployeeAPI = (*EmployeeService)(nil)

func parseOptionalUUID(s string) *uuid.UUID {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

func (s *EmployeeService) Create(ctx context.Context, userID, fullName, position, department, phone string, active bool) (*domain.Employee, error) {
	now := time.Now().UTC()
	e := &domain.Employee{
		ID:         uuid.New(),
		UserID:     parseOptionalUUID(userID),
		FullName:   strings.TrimSpace(fullName),
		Position:   strings.TrimSpace(position),
		Department: strings.TrimSpace(department),
		Phone:      phone,
		Active:     active,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *EmployeeService) Get(ctx context.Context, id string) (*domain.Employee, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	e, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return e, nil
}

func (s *EmployeeService) GetByUserID(ctx context.Context, userID string) (*domain.Employee, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrNotFound
	}
	e, err := s.repo.GetByUserID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return e, nil
}

func (s *EmployeeService) ResolveRef(ctx context.Context, ref string) (*domain.Employee, error) {
	uid, err := uuid.Parse(ref)
	if err != nil {
		return nil, ErrNotFound
	}
	e, err := s.repo.ResolveRef(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return e, nil
}

func (s *EmployeeService) List(ctx context.Context, limit, offset int32, search, position string, activeOnly bool) ([]*domain.Employee, int32, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, limit, offset, search, position, activeOnly)
}

func (s *EmployeeService) Update(ctx context.Context, id string, userID, fullName, position, department, phone *string, active *bool) (*domain.Employee, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	e, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if userID != nil {
		e.UserID = parseOptionalUUID(*userID)
	}
	if fullName != nil {
		e.FullName = strings.TrimSpace(*fullName)
	}
	if position != nil {
		e.Position = strings.TrimSpace(*position)
	}
	if department != nil {
		e.Department = strings.TrimSpace(*department)
	}
	if phone != nil {
		e.Phone = *phone
	}
	if active != nil {
		e.Active = *active
	}
	e.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *EmployeeService) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrNotFound
	}
	return s.repo.Delete(ctx, uid)
}
