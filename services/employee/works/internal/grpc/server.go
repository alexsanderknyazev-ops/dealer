package grpc

import (
	"context"
	"errors"

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

func toProto(w *domain.Work) *worksv1.Work {
	if w == nil {
		return nil
	}
	return &worksv1.Work{
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
}

func (s *Server) CreateWork(ctx context.Context, req *worksv1.CreateWorkRequest) (*worksv1.CreateWorkResponse, error) {
	w, err := s.svc.Create(ctx, req.Code, req.Name, req.Category, req.LaborHours, req.UnitPrice, req.Notes)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &worksv1.CreateWorkResponse{Work: toProto(w)}, nil
}

func (s *Server) GetWork(ctx context.Context, req *worksv1.GetWorkRequest) (*worksv1.GetWorkResponse, error) {
	w, err := s.svc.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "work not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &worksv1.GetWorkResponse{Work: toProto(w)}, nil
}

func (s *Server) ListWorks(ctx context.Context, req *worksv1.ListWorksRequest) (*worksv1.ListWorksResponse, error) {
	list, total, err := s.svc.List(ctx, req.Limit, req.Offset, req.Search, req.Category)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*worksv1.Work, len(list))
	for i, w := range list {
		out[i] = toProto(w)
	}
	return &worksv1.ListWorksResponse{Works: out, Total: total}, nil
}

func (s *Server) UpdateWork(ctx context.Context, req *worksv1.UpdateWorkRequest) (*worksv1.UpdateWorkResponse, error) {
	var code, name, category, laborHours, unitPrice, notes *string
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
	w, err := s.svc.Update(ctx, req.Id, code, name, category, laborHours, unitPrice, notes)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "work not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &worksv1.UpdateWorkResponse{Work: toProto(w)}, nil
}

func (s *Server) DeleteWork(ctx context.Context, req *worksv1.DeleteWorkRequest) (*worksv1.DeleteWorkResponse, error) {
	if err := s.svc.Delete(ctx, req.Id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "work not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &worksv1.DeleteWorkResponse{}, nil
}
