package grpcauth

import (
	"context"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/dealer/dealer/pkg/grpclient"
)

type Config struct {
	JWTSecret  string
	WriteRoles []string
}

type claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func UnaryServerInterceptor(cfg Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		token, err := bearerFromContext(ctx)
		if err != nil {
			return nil, err
		}
		role, err := validateToken(cfg.JWTSecret, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "unauthorized")
		}
		if isWriteMethod(info.FullMethod) && !hasRole(role, cfg.WriteRoles) {
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}
		ctx = grpclient.WithAuthorization(ctx, "Bearer "+token)
		return handler(ctx, req)
	}
}

func bearerFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization")
	}
	token := strings.TrimPrefix(vals[0], "Bearer ")
	if token == "" {
		return "", status.Error(codes.Unauthenticated, "missing authorization")
	}
	return token, nil
}

func validateToken(secret, tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", err
	}
	cl, ok := token.Claims.(*claims)
	if !ok {
		return "", jwt.ErrTokenInvalidClaims
	}
	return cl.Role, nil
}

func isWriteMethod(fullMethod string) bool {
	return strings.Contains(fullMethod, "Create") ||
		strings.Contains(fullMethod, "Update") ||
		strings.Contains(fullMethod, "Delete") ||
		strings.Contains(fullMethod, "Link") ||
		strings.Contains(fullMethod, "Unlink")
}

func hasRole(role string, allowed []string) bool {
	for _, r := range allowed {
		if role == r {
			return true
		}
	}
	return false
}
