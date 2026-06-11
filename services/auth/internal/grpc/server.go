package grpc

import (
	"context"
	"strings"

	"github.com/dealer/dealer/auth-service/internal/service"
	authv1 "github.com/dealer/dealer/pkg/pb/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Server реализует auth.v1.AuthService.
type Server struct {
	authv1.UnimplementedAuthServiceServer
	svc *service.AuthService
}

// NewServer создаёт gRPC-сервер auth-микросервиса.
func NewServer(svc *service.AuthService) *Server {
	return &Server{svc: svc}
}

// Register регистрирует пользователя.
func (s *Server) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password required")
	}
	user, access, refresh, expiresAt, err := s.svc.Register(ctx, req.Email, req.Password, req.Name, req.Phone)
	if err != nil {
		if err == service.ErrUserExists {
			return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.RegisterResponse{
		UserId:       user.ID.String(),
		Email:        user.Email,
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

// Login выполняет вход.
func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password required")
	}
	user, access, refresh, expiresAt, err := s.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		if err == service.ErrBadCredentials {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.LoginResponse{
		UserId:       user.ID.String(),
		Email:        user.Email,
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

// Refresh обновляет access по refresh-токену.
func (s *Server) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token required")
	}
	access, refresh, expiresAt, err := s.svc.Refresh(ctx, req.RefreshToken)
	if err != nil {
		if err == service.ErrInvalidToken {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired refresh token")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.RefreshResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

// Logout инвалидирует refresh-токен.
func (s *Server) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	_ = s.svc.Logout(ctx, req.RefreshToken)
	return &authv1.LogoutResponse{}, nil
}

// Validate проверяет access-токен (из тела запроса или Authorization metadata для GET /api/me).
func (s *Server) Validate(ctx context.Context, req *authv1.ValidateRequest) (*authv1.ValidateResponse, error) {
	token := req.AccessToken
	if token == "" {
		token = bearerFromMetadata(ctx)
	}
	if token == "" {
		return &authv1.ValidateResponse{Valid: false}, nil
	}
	userID, email, valid := s.svc.Validate(ctx, token)
	return &authv1.ValidateResponse{
		UserId: userID,
		Email:  email,
		Valid:  valid,
	}, nil
}

func bearerFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return ""
	}
	return strings.TrimPrefix(strings.TrimSpace(vals[0]), "Bearer ")
}
