package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/brands/internal/domain"
	"github.com/dealer/dealer/services/brands/internal/repository"
)

var ErrNotFound = errors.New("brand not found")

type BrandService struct {
	repo *repository.BrandRepository
}

func NewBrandService(repo *repository.BrandRepository) *BrandService {
	return &BrandService{repo: repo}
}

func (s *BrandService) Create(ctx context.Context, name string) (*domain.Brand, error) {
	now := time.Now().UTC()
	b := &domain.Brand{ID: uuid.New(), Name: name, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *BrandService) Get(ctx context.Context, id string) (*domain.Brand, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	b, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return b, nil
}

func (s *BrandService) List(ctx context.Context, limit, offset int32, search string) ([]*domain.Brand, int32, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, limit, offset, search)
}

func (s *BrandService) Update(ctx context.Context, id string, name *string) (*domain.Brand, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	b, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if name != nil {
		b.Name = *name
	}
	b.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *BrandService) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrNotFound
	}
	return s.repo.Delete(ctx, uid)
}
