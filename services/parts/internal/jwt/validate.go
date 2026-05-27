package jwt

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func Validate(secret, tokenString string) (userID, email, role string, err error) {
	if tokenString == "" {
		return "", "", "", errors.New("missing token")
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", "", "", errors.New("invalid token")
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", "", "", errors.New("invalid claims")
	}
	return claims.UserID, claims.Email, claims.Role, nil
}
