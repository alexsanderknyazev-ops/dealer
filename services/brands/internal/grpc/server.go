package grpc

import (
	"context"
	"errors"

	"github.com/dealer/dealer/services/brands/internal/domain"
	"github.com/dealer/dealer/services/brands/internal/service"
	brandsv1 "github.com/dealer/dealer/pkg/pb/brands/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	brandsv1.UnimplementedBrandsServiceServer
	svc *service.BrandService
}

func NewServer(svc *service.BrandService) *Server {
	return &Server{svc: svc}
}

func toProto(b *domain.Brand) *brandsv1.Brand {
	if b == nil {
		return nil
	}
	return &brandsv1.Brand{
		Id:        b.ID.String(),
		Name:      b.Name,
		CreatedAt: b.CreatedAt.Unix(),
		UpdatedAt: b.UpdatedAt.Unix(),
	}
}

func (s *Server) CreateBrand(ctx context.Context, req *brandsv1.CreateBrandRequest) (*brandsv1.CreateBrandResponse, error) {
	b, err := s.svc.Create(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &brandsv1.CreateBrandResponse{Brand: toProto(b)}, nil
}

func (s *Server) GetBrand(ctx context.Context, req *brandsv1.GetBrandRequest) (*brandsv1.GetBrandResponse, error) {
	b, err := s.svc.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "brand not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &brandsv1.GetBrandResponse{Brand: toProto(b)}, nil
}

func (s *Server) ListBrands(ctx context.Context, req *brandsv1.ListBrandsRequest) (*brandsv1.ListBrandsResponse, error) {
	list, total, err := s.svc.List(ctx, req.Limit, req.Offset, req.Search)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*brandsv1.Brand, len(list))
	for i, b := range list {
		out[i] = toProto(b)
	}
	return &brandsv1.ListBrandsResponse{Brands: out, Total: total}, nil
}

func (s *Server) UpdateBrand(ctx context.Context, req *brandsv1.UpdateBrandRequest) (*brandsv1.UpdateBrandResponse, error) {
	var name *string
	if req.Name != nil {
		v := req.GetName()
		name = &v
	}
	b, err := s.svc.Update(ctx, req.Id, name)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "brand not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &brandsv1.UpdateBrandResponse{Brand: toProto(b)}, nil
}

func (s *Server) DeleteBrand(ctx context.Context, req *brandsv1.DeleteBrandRequest) (*brandsv1.DeleteBrandResponse, error) {
	if err := s.svc.Delete(ctx, req.Id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "brand not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &brandsv1.DeleteBrandResponse{}, nil
}
