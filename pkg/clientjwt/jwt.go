package clientjwt

import (
	"context"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const Role = "client"

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func UserID(ctx context.Context, secret string) (uuid.UUID, error) {
	token, err := bearerFromContext(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	claims, err := parse(secret, token)
	if err != nil {
		return uuid.Nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	if claims.Role != Role {
		return uuid.Nil, status.Error(codes.PermissionDenied, "client role required")
	}
	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	return uid, nil
}

func ValidateHTTP(secret, authHeader string) (uuid.UUID, error) {
	token := strings.TrimPrefix(strings.TrimSpace(authHeader), "Bearer ")
	if token == "" {
		return uuid.Nil, status.Error(codes.Unauthenticated, "missing authorization")
	}
	claims, err := parse(secret, token)
	if err != nil || claims.Role != Role {
		return uuid.Nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	return uuid.Parse(claims.UserID)
}

func bearerFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing authorization")
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization")
	}
	token := strings.TrimPrefix(strings.TrimSpace(vals[0]), "Bearer ")
	if token == "" {
		return "", status.Error(codes.Unauthenticated, "missing authorization")
	}
	return token, nil
}

func parse(secret, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
