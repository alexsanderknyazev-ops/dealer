package grpc

import (
	"context"
	"errors"

	"github.com/dealer/dealer/customers-service/internal/domain"
	"github.com/dealer/dealer/customers-service/internal/service"
	customersv1 "github.com/dealer/dealer/pkg/pb/customers/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	customersv1.UnimplementedCustomersServiceServer
	svc service.CustomerAPI
}

func NewServer(svc service.CustomerAPI) *Server {
	return &Server{svc: svc}
}

func toProto(c *domain.Customer) *customersv1.Customer {
	if c == nil {
		return nil
	}
	return &customersv1.Customer{
		Id:           c.ID.String(),
		Name:         c.Name,
		Email:        c.Email,
		Phone:        c.Phone,
		CustomerType: c.CustomerType,
		Inn:          c.INN,
		Address:      c.Address,
		Notes:        c.Notes,
		CreatedAt:    c.CreatedAt.Unix(),
		UpdatedAt:    c.UpdatedAt.Unix(),
	}
}

func (s *Server) CreateCustomer(ctx context.Context, req *customersv1.CreateCustomerRequest) (*customersv1.CreateCustomerResponse, error) {
	c, err := s.svc.Create(ctx, service.CreateCustomerInput{
		Name: req.Name, Email: req.Email, Phone: req.Phone, CustomerType: req.CustomerType,
		INN: req.Inn, Address: req.Address, Notes: req.Notes,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &customersv1.CreateCustomerResponse{Customer: toProto(c)}, nil
}

func (s *Server) GetCustomer(ctx context.Context, req *customersv1.GetCustomerRequest) (*customersv1.GetCustomerResponse, error) {
	c, err := s.svc.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "customer not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &customersv1.GetCustomerResponse{Customer: toProto(c)}, nil
}

func (s *Server) ListCustomers(ctx context.Context, req *customersv1.ListCustomersRequest) (*customersv1.ListCustomersResponse, error) {
	list, total, err := s.svc.List(ctx, domain.CustomerListParams{Limit: req.Limit, Offset: req.Offset, Search: req.Search})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*customersv1.Customer, len(list))
	for i, c := range list {
		out[i] = toProto(c)
	}
	return &customersv1.ListCustomersResponse{Customers: out, Total: total}, nil
}

func (s *Server) UpdateCustomer(ctx context.Context, req *customersv1.UpdateCustomerRequest) (*customersv1.UpdateCustomerResponse, error) {
	c, err := s.svc.Update(ctx, req.Id, service.UpdateCustomerInput{
		Name: req.Name, Email: req.Email, Phone: req.Phone, CustomerType: req.CustomerType,
		INN: req.Inn, Address: req.Address, Notes: req.Notes,
	})
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "customer not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &customersv1.UpdateCustomerResponse{Customer: toProto(c)}, nil
}

func (s *Server) DeleteCustomer(ctx context.Context, req *customersv1.DeleteCustomerRequest) (*customersv1.DeleteCustomerResponse, error) {
	if err := s.svc.Delete(ctx, req.Id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "customer not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &customersv1.DeleteCustomerResponse{}, nil
}
