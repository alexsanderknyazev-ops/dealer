package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	worksv1 "github.com/dealer/dealer/pkg/pb/works/v1"
	"github.com/dealer/dealer/services/works/internal/domain"
	"github.com/dealer/dealer/services/works/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	worksv1.UnimplementedWorksServiceServer
	svc service.WorkAPI
}

func NewServer(svc service.WorkAPI) *Server {
	return &Server{svc: svc}
}

func parseUUIDOpt(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &u
}

func folderToProto(f *domain.WorkFolder) *worksv1.WorkFolder {
	if f == nil {
		return nil
	}
	parentID := ""
	if f.ParentID != nil {
		parentID = f.ParentID.String()
	}
	return &worksv1.WorkFolder{
		Id:        f.ID.String(),
		Name:      f.Name,
		ParentId:  parentID,
		CreatedAt: f.CreatedAt.Unix(),
		UpdatedAt: f.UpdatedAt.Unix(),
	}
}

func toProto(w *domain.Work) *worksv1.Work {
	if w == nil {
		return nil
	}
	out := &worksv1.Work{
		Id:         w.ID.String(),
		Code:       w.Code,
		Name:       w.Name,
		Category:   w.Category,
		LaborHours: w.LaborHours,
		UnitPrice:  w.UnitPrice,
		Notes:      w.Notes,
		CreatedAt:  w.CreatedAt.Unix(),
		UpdatedAt:  w.UpdatedAt.Unix(),
	}
	if w.FolderID != nil {
		out.FolderId = w.FolderID.String()
	}
	return out
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound), errors.Is(err, service.ErrFolderNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (s *Server) CreateWork(ctx context.Context, req *worksv1.CreateWorkRequest) (*worksv1.CreateWorkResponse, error) {
	w, err := s.svc.Create(ctx, req.Code, req.Name, req.Category, req.LaborHours, req.UnitPrice, req.Notes, parseUUIDOpt(req.FolderId))
	if err != nil {
		return nil, mapErr(err)
	}
	return &worksv1.CreateWorkResponse{Work: toProto(w)}, nil
}

func (s *Server) GetWork(ctx context.Context, req *worksv1.GetWorkRequest) (*worksv1.GetWorkResponse, error) {
	w, err := s.svc.Get(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return &worksv1.GetWorkResponse{Work: toProto(w)}, nil
}

func (s *Server) ListWorks(ctx context.Context, req *worksv1.ListWorksRequest) (*worksv1.ListWorksResponse, error) {
	list, total, err := s.svc.List(ctx, req.Limit, req.Offset, req.Search, req.Category, req.FolderId)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*worksv1.Work, len(list))
	for i, w := range list {
		out[i] = toProto(w)
	}
	return &worksv1.ListWorksResponse{Works: out, Total: total}, nil
}

func (s *Server) UpdateWork(ctx context.Context, req *worksv1.UpdateWorkRequest) (*worksv1.UpdateWorkResponse, error) {
	var code, name, category, laborHours, unitPrice, notes, folderID *string
	if req.Code != nil {
		v := req.GetCode()
		code = &v
	}
	if req.Name != nil {
		v := req.GetName()
		name = &v
	}
	if req.Category != nil {
		v := req.GetCategory()
		category = &v
	}
	if req.LaborHours != nil {
		v := req.GetLaborHours()
		laborHours = &v
	}
	if req.UnitPrice != nil {
		v := req.GetUnitPrice()
		unitPrice = &v
	}
	if req.Notes != nil {
		v := req.GetNotes()
		notes = &v
	}
	if req.FolderId != nil {
		v := req.GetFolderId()
		folderID = &v
	}
	w, err := s.svc.Update(ctx, req.Id, code, name, category, laborHours, unitPrice, notes, folderID)
	if err != nil {
		return nil, mapErr(err)
	}
	return &worksv1.UpdateWorkResponse{Work: toProto(w)}, nil
}

func (s *Server) DeleteWork(ctx context.Context, req *worksv1.DeleteWorkRequest) (*worksv1.DeleteWorkResponse, error) {
	if err := s.svc.Delete(ctx, req.Id); err != nil {
		return nil, mapErr(err)
	}
	return &worksv1.DeleteWorkResponse{}, nil
}

func (s *Server) CreateFolder(ctx context.Context, req *worksv1.CreateFolderRequest) (*worksv1.CreateFolderResponse, error) {
	f, err := s.svc.CreateFolder(ctx, req.Name, parseUUIDOpt(req.ParentId))
	if err != nil {
		return nil, mapErr(err)
	}
	return &worksv1.CreateFolderResponse{Folder: folderToProto(f)}, nil
}

func (s *Server) GetFolder(ctx context.Context, req *worksv1.GetFolderRequest) (*worksv1.GetFolderResponse, error) {
	f, err := s.svc.GetFolder(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return &worksv1.GetFolderResponse{Folder: folderToProto(f)}, nil
}

func (s *Server) ListFolders(ctx context.Context, req *worksv1.ListFoldersRequest) (*worksv1.ListFoldersResponse, error) {
	list, err := s.svc.ListFolders(ctx, parseUUIDOpt(req.ParentId))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*worksv1.WorkFolder, len(list))
	for i, f := range list {
		out[i] = folderToProto(f)
	}
	return &worksv1.ListFoldersResponse{Folders: out}, nil
}

func (s *Server) UpdateFolder(ctx context.Context, req *worksv1.UpdateFolderRequest) (*worksv1.UpdateFolderResponse, error) {
	var name *string
	if req.Name != nil {
		v := req.GetName()
		name = &v
	}
	var parentIDOpt *string
	if req.ParentId != nil {
		v := req.GetParentId()
		parentIDOpt = &v
	}
	f, err := s.svc.UpdateFolder(ctx, req.Id, name, parentIDOpt)
	if err != nil {
		return nil, mapErr(err)
	}
	return &worksv1.UpdateFolderResponse{Folder: folderToProto(f)}, nil
}

func (s *Server) DeleteFolder(ctx context.Context, req *worksv1.DeleteFolderRequest) (*worksv1.DeleteFolderResponse, error) {
	if err := s.svc.DeleteFolder(ctx, req.Id); err != nil {
		return nil, mapErr(err)
	}
	return &worksv1.DeleteFolderResponse{}, nil
}
