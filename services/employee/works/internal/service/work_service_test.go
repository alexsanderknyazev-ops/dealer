package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dealer/dealer/services/works/internal/domain"
)

type memWorkRepo struct {
	byID map[uuid.UUID]*domain.Work
	err  error
}

func (m *memWorkRepo) Create(_ context.Context, w *domain.Work) error {
	if m.err != nil {
		return m.err
	}
	if m.byID == nil {
		m.byID = make(map[uuid.UUID]*domain.Work)
	}
	cp := *w
	m.byID[w.ID] = &cp
	return nil
}

func (m *memWorkRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Work, error) {
	if m.err != nil {
		return nil, m.err
	}
	w, ok := m.byID[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *w
	return &cp, nil
}

func (m *memWorkRepo) List(_ context.Context, _, _ int32, _, _ string, folderID *uuid.UUID) ([]*domain.Work, int32, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	var out []*domain.Work
	for _, w := range m.byID {
		if folderID != nil {
			if w.FolderID == nil || *w.FolderID != *folderID {
				continue
			}
		}
		cp := *w
		out = append(out, &cp)
	}
	return out, int32(len(out)), nil
}

func (m *memWorkRepo) Update(_ context.Context, w *domain.Work) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.byID[w.ID]; !ok {
		return pgx.ErrNoRows
	}
	cp := *w
	m.byID[w.ID] = &cp
	return nil
}

func (m *memWorkRepo) Delete(_ context.Context, id uuid.UUID) error {
	if m.err != nil {
		return m.err
	}
	delete(m.byID, id)
	return nil
}

type memFolderRepo struct {
	folders map[uuid.UUID]*domain.WorkFolder
}

func (f *memFolderRepo) Create(_ context.Context, folder *domain.WorkFolder) error {
	if f.folders == nil {
		f.folders = make(map[uuid.UUID]*domain.WorkFolder)
	}
	cp := *folder
	f.folders[folder.ID] = &cp
	return nil
}

func (f *memFolderRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.WorkFolder, error) {
	folder, ok := f.folders[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *folder
	return &cp, nil
}

func (f *memFolderRepo) ListByParent(context.Context, *uuid.UUID) ([]*domain.WorkFolder, error) {
	return nil, nil
}

func (f *memFolderRepo) Update(_ context.Context, folder *domain.WorkFolder) error {
	f.folders[folder.ID] = folder
	return nil
}

func (f *memFolderRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(f.folders, id)
	return nil
}

func TestWorkService_Create_Defaults(t *testing.T) {
	s := NewWorkService(&memWorkRepo{byID: map[uuid.UUID]*domain.Work{}}, &memFolderRepo{})
	w, err := s.Create(context.Background(), "LAB-100", "Test work", "TO", "", "", "notes", nil)
	if err != nil {
		t.Fatal(err)
	}
	if w.LaborHours != "1" || w.UnitPrice != "0" || w.Code != "LAB-100" {
		t.Fatalf("%+v", w)
	}
}

func TestWorkService_Get_NotFound(t *testing.T) {
	s := NewWorkService(&memWorkRepo{byID: map[uuid.UUID]*domain.Work{}}, &memFolderRepo{})
	_, err := s.Get(context.Background(), uuid.New().String())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v", err)
	}
	_, err = s.Get(context.Background(), "bad")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestWorkService_Update(t *testing.T) {
	repo := &memWorkRepo{byID: map[uuid.UUID]*domain.Work{}}
	s := NewWorkService(repo, &memFolderRepo{})
	w, _ := s.Create(context.Background(), "LAB-1", "Name", "Cat", "1", "100", "", nil)

	newName := "Updated"
	updated, err := s.Update(context.Background(), w.ID.String(), nil, &newName, nil, nil, nil, nil, nil)
	if err != nil || updated.Name != "Updated" {
		t.Fatalf("err=%v name=%q", err, updated.Name)
	}
}

func TestWorkService_List_DefaultLimit(t *testing.T) {
	s := NewWorkService(&memWorkRepo{byID: map[uuid.UUID]*domain.Work{}}, &memFolderRepo{})
	_, _, err := s.List(context.Background(), 0, 0, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestWorkService_Delete(t *testing.T) {
	repo := &memWorkRepo{byID: map[uuid.UUID]*domain.Work{}}
	s := NewWorkService(repo, &memFolderRepo{})
	w, _ := s.Create(context.Background(), "LAB-2", "X", "", "1", "0", "", nil)
	if err := s.Delete(context.Background(), w.ID.String()); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(context.Background(), "bad"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestWorkService_CreateFolder(t *testing.T) {
	s := NewWorkService(&memWorkRepo{byID: map[uuid.UUID]*domain.Work{}}, &memFolderRepo{folders: map[uuid.UUID]*domain.WorkFolder{}})
	f, err := s.CreateFolder(context.Background(), "ТО", nil)
	if err != nil || f.Name != "ТО" {
		t.Fatalf("err=%v folder=%+v", err, f)
	}
}
