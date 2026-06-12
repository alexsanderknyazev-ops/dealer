package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/dealer/dealer/client-registration-service/internal/domain"
	"github.com/dealer/dealer/client-registration-service/internal/service"
	"github.com/dealer/dealer/pkg/clientjwt"
	clientsv1 "github.com/dealer/dealer/pkg/pb/clients/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	clientsv1.UnimplementedClientRegistrationPublicServiceServer
	clientsv1.UnimplementedClientAccountServiceServer
	svc       service.ClientAPI
	jwtSecret string
}

func NewServer(svc service.ClientAPI, jwtSecret string) *Server {
	return &Server{svc: svc, jwtSecret: jwtSecret}
}

func (s *Server) RegisterClient(ctx context.Context, req *clientsv1.RegisterClientRequest) (*clientsv1.RegisterClientResponse, error) {
	if req.Email == "" || req.Password == "" || req.FullName == "" || req.Phone == "" || req.Vin == "" {
		return nil, status.Error(codes.InvalidArgument, "email, password, full_name, phone and vin required")
	}
	result, err := s.svc.Register(ctx, req.Email, req.Password, req.FullName, req.Phone, req.Vin)
	if err != nil {
		return nil, mapErr(err)
	}
	return &clientsv1.RegisterClientResponse{
		ClientId:     result.Client.ID.String(),
		UserId:       result.Client.UserID.String(),
		Email:        result.Client.Email,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt,
	}, nil
}

func (s *Server) AddVehicle(ctx context.Context, req *clientsv1.AddVehicleRequest) (*clientsv1.AddVehicleResponse, error) {
	if req.Vin == "" {
		return nil, status.Error(codes.InvalidArgument, "vin required")
	}
	userID, err := clientjwt.UserID(ctx, s.jwtSecret)
	if err != nil {
		return nil, err
	}
	v, err := s.svc.AddVehicle(ctx, userID, req.Vin)
	if err != nil {
		return nil, mapErr(err)
	}
	return &clientsv1.AddVehicleResponse{Vehicle: toProtoVehicle(v)}, nil
}

func (s *Server) ListMyVehicles(ctx context.Context, _ *clientsv1.ListMyVehiclesRequest) (*clientsv1.ListMyVehiclesResponse, error) {
	userID, err := clientjwt.UserID(ctx, s.jwtSecret)
	if err != nil {
		return nil, err
	}
	list, err := s.svc.ListVehicles(ctx, userID)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*clientsv1.ClientVehicle, len(list))
	for i, v := range list {
		out[i] = toProtoVehicle(v)
	}
	return &clientsv1.ListMyVehiclesResponse{Vehicles: out}, nil
}

func (s *Server) GetProfile(ctx context.Context, _ *clientsv1.GetProfileRequest) (*clientsv1.GetProfileResponse, error) {
	userID, err := clientjwt.UserID(ctx, s.jwtSecret)
	if err != nil {
		return nil, err
	}
	client, vehicles, err := s.svc.GetProfile(ctx, userID)
	if err != nil {
		return nil, mapErr(err)
	}
	return &clientsv1.GetProfileResponse{Profile: toProtoProfile(client, vehicles)}, nil
}

func (s *Server) ListClientNotifications(ctx context.Context, _ *clientsv1.ListClientNotificationsRequest) (*clientsv1.ListClientNotificationsResponse, error) {
	userID, err := clientjwt.UserID(ctx, s.jwtSecret)
	if err != nil {
		return nil, err
	}
	list, err := s.svc.ListNotifications(ctx, userID)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*clientsv1.ClientNotification, len(list))
	for i, n := range list {
		out[i] = toProtoNotification(n)
	}
	return &clientsv1.ListClientNotificationsResponse{Notifications: out}, nil
}

func (s *Server) DismissClientNotification(ctx context.Context, req *clientsv1.DismissClientNotificationRequest) (*clientsv1.DismissClientNotificationResponse, error) {
	userID, err := clientjwt.UserID(ctx, s.jwtSecret)
	if err != nil {
		return nil, err
	}
	nid, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	if err := s.svc.DismissNotification(ctx, userID, nid); err != nil {
		return nil, mapErr(err)
	}
	return &clientsv1.DismissClientNotificationResponse{}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, service.ErrVINNotFound):
		return status.Error(codes.NotFound, "vehicle with this vin not found")
	case errors.Is(err, service.ErrVINAlreadyLinked):
		return status.Error(codes.AlreadyExists, "vin already linked to another client")
	case errors.Is(err, service.ErrUserExists):
		return status.Error(codes.AlreadyExists, "user with this email already exists")
	case errors.Is(err, service.ErrAuthNotReady):
		return status.Error(codes.Unavailable, "client auth not ready, try login shortly")
	case errors.Is(err, service.ErrClientNotFound):
		return status.Error(codes.NotFound, "client not found")
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func toProtoVehicle(v *domain.ClientVehicle) *clientsv1.ClientVehicle {
	if v == nil {
		return nil
	}
	return &clientsv1.ClientVehicle{
		Id:        v.ID.String(),
		VehicleId: v.VehicleID.String(),
		Vin:       v.VIN,
		Make:      v.Make,
		Model:     v.Model,
		Year:      v.Year,
		AddedAt:   v.AddedAt.Unix(),
	}
}

func toProtoNotification(n *domain.ClientNotification) *clientsv1.ClientNotification {
	if n == nil {
		return nil
	}
	return &clientsv1.ClientNotification{
		Id: n.ID.String(), Kind: n.Kind, SourceType: n.SourceType, SourceId: n.SourceID.String(),
		Title: n.Title, Body: n.Body, Status: n.Status, CreatedAt: n.CreatedAt.Unix(),
	}
}

func toProtoProfile(c *domain.Client, vehicles []*domain.ClientVehicle) *clientsv1.ClientProfile {
	if c == nil {
		return nil
	}
	out := &clientsv1.ClientProfile{
		Id:        c.ID.String(),
		UserId:    c.UserID.String(),
		Email:     c.Email,
		FullName:  c.FullName,
		Phone:     c.Phone,
		CreatedAt: c.CreatedAt.Unix(),
	}
	for _, v := range vehicles {
		out.Vehicles = append(out.Vehicles, toProtoVehicle(v))
	}
	return out
}
