package grpc

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/dealer/dealer/client-auth-service/internal/service"
	clientauthv1 "github.com/dealer/dealer/pkg/pb/clientauth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	clientauthv1.UnimplementedClientAuthServiceServer
	clientauthv1.UnimplementedClientAuthPublicServiceServer
	clientauthv1.UnimplementedClientAuthSessionServiceServer
	svc *service.AuthService
}

func NewServer(svc *service.AuthService) *Server {
	return &Server{svc: svc}
}

func (s *Server) Login(ctx context.Context, req *clientauthv1.LoginRequest) (*clientauthv1.LoginResponse, error) {
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
	return &clientauthv1.LoginResponse{
		UserId:       user.ID.String(),
		Email:        user.Email,
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

func (s *Server) Refresh(ctx context.Context, req *clientauthv1.RefreshRequest) (*clientauthv1.RefreshResponse, error) {
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
	return &clientauthv1.RefreshResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *clientauthv1.LogoutRequest) (*clientauthv1.LogoutResponse, error) {
	_ = s.svc.Logout(ctx, req.RefreshToken)
	return &clientauthv1.LogoutResponse{}, nil
}

func (s *Server) Validate(ctx context.Context, req *clientauthv1.ValidateRequest) (*clientauthv1.ValidateResponse, error) {
	token := req.AccessToken
	if token == "" {
		token = bearerFromMetadata(ctx)
	}
	if token == "" {
		return &clientauthv1.ValidateResponse{Valid: false}, nil
	}
	userID, email, valid := s.svc.Validate(ctx, token)
	return &clientauthv1.ValidateResponse{
		UserId: userID,
		Email:  email,
		Valid:  valid,
	}, nil
}

func (s *Server) IssueTokens(ctx context.Context, req *clientauthv1.IssueTokensRequest) (*clientauthv1.IssueTokensResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	access, refresh, expiresAt, err := s.svc.IssueTokens(ctx, userID)
	if err != nil {
		if err == service.ErrNotClient {
			return nil, status.Error(codes.NotFound, "client user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &clientauthv1.IssueTokensResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    expiresAt.Unix(),
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
