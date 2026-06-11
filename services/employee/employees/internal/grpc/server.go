package grpc

import (
	"context"
	"errors"

	employeesv1 "github.com/dealer/dealer/pkg/pb/employees/v1"
	"github.com/dealer/dealer/services/employees/internal/domain"
	"github.com/dealer/dealer/services/employees/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	employeesv1.UnimplementedEmployeesServiceServer
	svc service.EmployeeAPI
}

func NewServer(svc service.EmployeeAPI) *Server {
	return &Server{svc: svc}
}

func toProto(e *domain.Employee) *employeesv1.Employee {
	if e == nil {
		return nil
	}
	out := &employeesv1.Employee{
		Id:         e.ID.String(),
		FullName:   e.FullName,
		Position:   e.Position,
		Department: e.Department,
		Phone:      e.Phone,
		Active:     e.Active,
		CreatedAt:  e.CreatedAt.Unix(),
		UpdatedAt:  e.UpdatedAt.Unix(),
	}
	if e.UserID != nil {
		out.UserId = e.UserID.String()
	}
	return out
}

func (s *Server) CreateEmployee(ctx context.Context, req *employeesv1.CreateEmployeeRequest) (*employeesv1.CreateEmployeeResponse, error) {
	e, err := s.svc.Create(ctx, req.UserId, req.FullName, req.Position, req.Department, req.Phone, req.Active)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &employeesv1.CreateEmployeeResponse{Employee: toProto(e)}, nil
}

func (s *Server) GetEmployee(ctx context.Context, req *employeesv1.GetEmployeeRequest) (*employeesv1.GetEmployeeResponse, error) {
	e, err := s.svc.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "employee not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &employeesv1.GetEmployeeResponse{Employee: toProto(e)}, nil
}

func (s *Server) GetEmployeeByUserID(ctx context.Context, req *employeesv1.GetEmployeeByUserIDRequest) (*employeesv1.GetEmployeeResponse, error) {
	e, err := s.svc.GetByUserID(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "employee not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &employeesv1.GetEmployeeResponse{Employee: toProto(e)}, nil
}

func (s *Server) ListEmployees(ctx context.Context, req *employeesv1.ListEmployeesRequest) (*employeesv1.ListEmployeesResponse, error) {
	list, total, err := s.svc.List(ctx, req.Limit, req.Offset, req.Search, req.Position, req.ActiveOnly)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*employeesv1.Employee, len(list))
	for i, e := range list {
		out[i] = toProto(e)
	}
	return &employeesv1.ListEmployeesResponse{Employees: out, Total: total}, nil
}

func (s *Server) UpdateEmployee(ctx context.Context, req *employeesv1.UpdateEmployeeRequest) (*employeesv1.UpdateEmployeeResponse, error) {
	var userID, fullName, position, department, phone *string
	if req.UserId != nil {
		v := req.GetUserId()
		userID = &v
	}
	if req.FullName != nil {
		v := req.GetFullName()
		fullName = &v
	}
	if req.Position != nil {
		v := req.GetPosition()
		position = &v
	}
	if req.Department != nil {
		v := req.GetDepartment()
		department = &v
	}
	if req.Phone != nil {
		v := req.GetPhone()
		phone = &v
	}
	var active *bool
	if req.Active != nil {
		v := req.GetActive()
		active = &v
	}
	e, err := s.svc.Update(ctx, req.Id, userID, fullName, position, department, phone, active)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "employee not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &employeesv1.UpdateEmployeeResponse{Employee: toProto(e)}, nil
}

func (s *Server) DeleteEmployee(ctx context.Context, req *employeesv1.DeleteEmployeeRequest) (*employeesv1.DeleteEmployeeResponse, error) {
	if err := s.svc.Delete(ctx, req.Id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "employee not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &employeesv1.DeleteEmployeeResponse{}, nil
}
