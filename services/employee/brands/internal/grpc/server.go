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
	svc   service.BrandAPI
	rates service.LaborRateAPI
}

func NewServer(svc service.BrandAPI, rates service.LaborRateAPI) *Server {
	return &Server{svc: svc, rates: rates}
}

func toProtoLaborRate(r *domain.BrandLaborRate) *brandsv1.BrandLaborRate {
	if r == nil {
		return nil
	}
	return &brandsv1.BrandLaborRate{
		Id:                    r.ID.String(),
		BrandId:               r.BrandID.String(),
		DealerPointId:         r.DealerPointID.String(),
		WarrantyHourPrice:     r.WarrantyHourPrice,
		CommercialHourPrice:   r.CommercialHourPrice,
		CreatedAt:             r.CreatedAt.Unix(),
		UpdatedAt:             r.UpdatedAt.Unix(),
	}
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

func (s *Server) ListBrandLaborRates(ctx context.Context, req *brandsv1.ListBrandLaborRatesRequest) (*brandsv1.ListBrandLaborRatesResponse, error) {
	if s.rates == nil {
		return &brandsv1.ListBrandLaborRatesResponse{}, nil
	}
	list, total, err := s.rates.List(ctx, req.Limit, req.Offset, req.BrandId, req.DealerPointId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*brandsv1.BrandLaborRate, len(list))
	for i, r := range list {
		out[i] = toProtoLaborRate(r)
	}
	return &brandsv1.ListBrandLaborRatesResponse{BrandLaborRates: out, Total: total}, nil
}

func (s *Server) UpdateBrandLaborRate(ctx context.Context, req *brandsv1.UpdateBrandLaborRateRequest) (*brandsv1.UpdateBrandLaborRateResponse, error) {
	if s.rates == nil {
		return nil, status.Error(codes.Unavailable, "labor rates unavailable")
	}
	r, err := s.rates.Update(ctx, req.BrandId, req.DealerPointId, req.WarrantyHourPrice, req.CommercialHourPrice)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "brand not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &brandsv1.UpdateBrandLaborRateResponse{BrandLaborRate: toProtoLaborRate(r)}, nil
}

func (s *Server) DeleteBrandLaborRate(ctx context.Context, req *brandsv1.DeleteBrandLaborRateRequest) (*brandsv1.DeleteBrandLaborRateResponse, error) {
	if s.rates == nil {
		return nil, status.Error(codes.Unavailable, "labor rates unavailable")
	}
	if err := s.rates.Delete(ctx, req.Id); err != nil {
		if errors.Is(err, service.ErrLaborRateNotFound) {
			return nil, status.Error(codes.NotFound, "brand labor rate not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &brandsv1.DeleteBrandLaborRateResponse{}, nil
}

func (s *Server) ResolveBrandLaborRate(ctx context.Context, req *brandsv1.ResolveBrandLaborRateRequest) (*brandsv1.ResolveBrandLaborRateResponse, error) {
	if s.rates == nil {
		return &brandsv1.ResolveBrandLaborRateResponse{}, nil
	}
	warranty, commercial, hour, found, err := s.rates.Resolve(ctx, req.BrandId, req.DealerPointId, req.RepairType)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &brandsv1.ResolveBrandLaborRateResponse{
		WarrantyHourPrice:   warranty,
		CommercialHourPrice: commercial,
		HourPrice:           hour,
		Found:               found,
	}, nil
}
