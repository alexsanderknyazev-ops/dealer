package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/works/internal/domain"
)

var ErrNotFound = errors.New("work not found")

type WorkAPI interface {
	Create(ctx context.Context, code, name, category, laborHours, unitPrice, notes string) (*domain.Work, error)
	Get(ctx context.Context, id string) (*domain.Work, error)
	List(ctx context.Context, limit, offset int32, search, category string) ([]*domain.Work, int32, error)
	Update(ctx context.Context, id string, code, name, category, laborHours, unitPrice, notes *string) (*domain.Work, error)
	Delete(ctx context.Context, id string) error
}

type workRepository interface {
	Create(ctx context.Context, w *domain.Work) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Work, error)
	List(ctx context.Context, limit, offset int32, search, category string) ([]*domain.Work, int32, error)
	Update(ctx context.Context, w *domain.Work) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type WorkService struct {
	repo workRepository
}

func NewWorkService(repo workRepository) *WorkService {
	return &WorkService{repo: repo}
}

var _ WorkAPI = (*WorkService)(nil)

func defaultLaborHours(v string) string {
	if strings.TrimSpace(v) == "" {
		return "1"
	}
	return v
}

func defaultUnitPrice(v string) string {
	if strings.TrimSpace(v) == "" {
		return "0"
	}
	return v
}

func (s *WorkService) Create(ctx context.Context, code, name, category, laborHours, unitPrice, notes string) (*domain.Work, error) {
	now := time.Now().UTC()
	w := &domain.Work{
		ID:         uuid.New(),
		Code:       strings.TrimSpace(code),
		Name:       strings.TrimSpace(name),
		Category:   strings.TrimSpace(category),
		LaborHours: defaultLaborHours(laborHours),
		UnitPrice:  defaultUnitPrice(unitPrice),
		Notes:      notes,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.Create(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *WorkService) Get(ctx context.Context, id string) (*domain.Work, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	w, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return w, nil
}

func (s *WorkService) List(ctx context.Context, limit, offset int32, search, category string) ([]*domain.Work, int32, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, limit, offset, search, category)
}

func (s *WorkService) Update(ctx context.Context, id string, code, name, category, laborHours, unitPrice, notes *string) (*domain.Work, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	w, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if code != nil {
		w.Code = strings.TrimSpace(*code)
	}
	if name != nil {
		w.Name = strings.TrimSpace(*name)
	}
	if category != nil {
		w.Category = strings.TrimSpace(*category)
	}
	if laborHours != nil {
		w.LaborHours = defaultLaborHours(*laborHours)
	}
	if unitPrice != nil {
		w.UnitPrice = defaultUnitPrice(*unitPrice)
	}
	if notes != nil {
		w.Notes = *notes
	}
	w.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *WorkService) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrNotFound
	}
	return s.repo.Delete(ctx, uid)
}
