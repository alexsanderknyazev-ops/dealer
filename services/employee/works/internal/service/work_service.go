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

var (
	ErrNotFound       = errors.New("work not found")
	ErrFolderNotFound = errors.New("folder not found")
)

type WorkAPI interface {
	Create(ctx context.Context, code, name, category, laborHours, unitPrice, notes string, folderID *uuid.UUID) (*domain.Work, error)
	Get(ctx context.Context, id string) (*domain.Work, error)
	List(ctx context.Context, limit, offset int32, search, category, folderID string) ([]*domain.Work, int32, error)
	Update(ctx context.Context, id string, code, name, category, laborHours, unitPrice, notes, folderID *string) (*domain.Work, error)
	Delete(ctx context.Context, id string) error
	CreateFolder(ctx context.Context, name string, parentID *uuid.UUID) (*domain.WorkFolder, error)
	GetFolder(ctx context.Context, id string) (*domain.WorkFolder, error)
	ListFolders(ctx context.Context, parentID *uuid.UUID) ([]*domain.WorkFolder, error)
	UpdateFolder(ctx context.Context, id string, name *string, parentIDOpt *string) (*domain.WorkFolder, error)
	DeleteFolder(ctx context.Context, id string) error
}

type workRepository interface {
	Create(ctx context.Context, w *domain.Work) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Work, error)
	List(ctx context.Context, limit, offset int32, search, category string, folderID *uuid.UUID) ([]*domain.Work, int32, error)
	Update(ctx context.Context, w *domain.Work) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type folderRepository interface {
	Create(ctx context.Context, f *domain.WorkFolder) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkFolder, error)
	ListByParent(ctx context.Context, parentID *uuid.UUID) ([]*domain.WorkFolder, error)
	Update(ctx context.Context, f *domain.WorkFolder) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type WorkService struct {
	repo       workRepository
	folderRepo folderRepository
}

func NewWorkService(repo workRepository, folderRepo folderRepository) *WorkService {
	return &WorkService{repo: repo, folderRepo: folderRepo}
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

func parseFolderFilter(folderID string) (*uuid.UUID, error) {
	if strings.TrimSpace(folderID) == "" {
		return nil, nil
	}
	uid, err := uuid.Parse(folderID)
	if err != nil {
		return nil, err
	}
	return &uid, nil
}

func (s *WorkService) validateFolder(ctx context.Context, folderID *uuid.UUID) error {
	if folderID == nil {
		return nil
	}
	_, err := s.folderRepo.GetByID(ctx, *folderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrFolderNotFound
		}
		return err
	}
	return nil
}

func (s *WorkService) Create(ctx context.Context, code, name, category, laborHours, unitPrice, notes string, folderID *uuid.UUID) (*domain.Work, error) {
	if err := s.validateFolder(ctx, folderID); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	w := &domain.Work{
		ID:         uuid.New(),
		Code:       strings.TrimSpace(code),
		Name:       strings.TrimSpace(name),
		Category:   strings.TrimSpace(category),
		FolderID:   folderID,
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

func (s *WorkService) List(ctx context.Context, limit, offset int32, search, category, folderID string) ([]*domain.Work, int32, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	fid, err := parseFolderFilter(folderID)
	if err != nil {
		return nil, 0, ErrFolderNotFound
	}
	return s.repo.List(ctx, limit, offset, search, category, fid)
}

func (s *WorkService) Update(ctx context.Context, id string, code, name, category, laborHours, unitPrice, notes, folderID *string) (*domain.Work, error) {
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
	if folderID != nil {
		if *folderID == "" {
			w.FolderID = nil
		} else {
			fid, err := uuid.Parse(*folderID)
			if err != nil {
				return nil, ErrFolderNotFound
			}
			if err := s.validateFolder(ctx, &fid); err != nil {
				return nil, err
			}
			w.FolderID = &fid
		}
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

func (s *WorkService) CreateFolder(ctx context.Context, name string, parentID *uuid.UUID) (*domain.WorkFolder, error) {
	now := time.Now().UTC()
	f := &domain.WorkFolder{
		ID:        uuid.New(),
		Name:      name,
		ParentID:  parentID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.folderRepo.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *WorkService) GetFolder(ctx context.Context, id string) (*domain.WorkFolder, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrFolderNotFound
	}
	f, err := s.folderRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFolderNotFound
		}
		return nil, err
	}
	return f, nil
}

func (s *WorkService) ListFolders(ctx context.Context, parentID *uuid.UUID) ([]*domain.WorkFolder, error) {
	return s.folderRepo.ListByParent(ctx, parentID)
}

func (s *WorkService) UpdateFolder(ctx context.Context, id string, name *string, parentIDOpt *string) (*domain.WorkFolder, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrFolderNotFound
	}
	existing, err := s.folderRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, ErrFolderNotFound
	}
	if name != nil {
		existing.Name = *name
	}
	if parentIDOpt != nil {
		if *parentIDOpt == "" {
			existing.ParentID = nil
		} else {
			pid, err := uuid.Parse(*parentIDOpt)
			if err == nil {
				existing.ParentID = &pid
			}
		}
	}
	existing.UpdatedAt = time.Now().UTC()
	if err := s.folderRepo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *WorkService) DeleteFolder(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrFolderNotFound
	}
	return s.folderRepo.Delete(ctx, uid)
}
