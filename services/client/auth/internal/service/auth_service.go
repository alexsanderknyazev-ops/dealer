package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/dealer/dealer/client-auth-service/internal/domain"
)

var (
	ErrBadCredentials = errors.New("invalid email or password")
	ErrInvalidToken   = errors.New("invalid or expired token")
	ErrNotClient      = errors.New("user is not a client")
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type AuthConfig struct {
	JWTSecret     string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	RefreshPrefix string
}

type userRepository interface {
	Create(ctx context.Context, u *domain.User) error
	EnsureExists(ctx context.Context, u *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type AuthService struct {
	repo userRepository
	rdb  *redis.Client
	cfg  AuthConfig
}

func NewAuthService(repo userRepository, rdb *redis.Client, cfg AuthConfig) *AuthService {
	return &AuthService{repo: repo, rdb: rdb, cfg: cfg}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, string, string, time.Time, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", time.Time{}, ErrBadCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", time.Time{}, ErrBadCredentials
	}
	return s.issueSession(ctx, user)
}

func (s *AuthService) RegisterFromEvent(ctx context.Context, userID uuid.UUID, email, passwordHash, fullName, phone string) error {
	u := &domain.User{
		ID:           userID,
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Phone:        phone,
		Role:         domain.ClientRole,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	return s.repo.EnsureExists(ctx, u)
}

func (s *AuthService) IssueTokens(ctx context.Context, userID uuid.UUID) (string, string, time.Time, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", time.Time{}, ErrNotClient
		}
		return "", "", time.Time{}, err
	}
	_, access, refresh, expiresAt, err := s.issueSession(ctx, user)
	if err != nil {
		return "", "", time.Time{}, err
	}
	return access, refresh, expiresAt, nil
}

func (s *AuthService) issueSession(ctx context.Context, user *domain.User) (*domain.User, string, string, time.Time, error) {
	accessToken, expiresAt, err := s.issueAccessToken(user)
	if err != nil {
		return user, "", "", time.Time{}, err
	}
	refreshToken, err := s.issueRefreshToken(ctx, user)
	if err != nil {
		return user, "", "", time.Time{}, err
	}
	return user, accessToken, refreshToken, expiresAt, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (accessToken, newRefresh string, expiresAt time.Time, err error) {
	userID, err := s.validateRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", time.Time{}, err
	}
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return "", "", time.Time{}, ErrInvalidToken
	}
	accessToken, expiresAt, err = s.issueAccessToken(user)
	if err != nil {
		return "", "", time.Time{}, err
	}
	_ = s.revokeRefreshToken(ctx, refreshToken)
	newRefresh, err = s.issueRefreshToken(ctx, user)
	if err != nil {
		return accessToken, "", expiresAt, err
	}
	return accessToken, newRefresh, expiresAt, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.revokeRefreshToken(ctx, refreshToken)
}

func (s *AuthService) Validate(ctx context.Context, accessToken string) (userID, email string, valid bool) {
	claims, err := s.parseAccessToken(accessToken)
	if err != nil {
		return "", "", false
	}
	if claims.Role != domain.ClientRole {
		return "", "", false
	}
	return claims.UserID, claims.Email, true
}

func (s *AuthService) issueAccessToken(u *domain.User) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(s.cfg.AccessTTL)
	claims := &JWTClaims{
		UserID: u.ID.String(),
		Email:  u.Email,
		Role:   u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (s *AuthService) issueRefreshToken(ctx context.Context, u *domain.User) (string, error) {
	refreshToken := uuid.New().String()
	key := s.cfg.RefreshPrefix + refreshToken
	payload := map[string]string{"user_id": u.ID.String(), "email": u.Email}
	data, _ := json.Marshal(payload)
	if err := s.rdb.Set(ctx, key, data, s.cfg.RefreshTTL).Err(); err != nil {
		return "", err
	}
	return refreshToken, nil
}

func (s *AuthService) validateRefreshToken(ctx context.Context, refreshToken string) (uuid.UUID, error) {
	key := s.cfg.RefreshPrefix + refreshToken
	data, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	var payload struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	return uuid.Parse(payload.UserID)
}

func (s *AuthService) revokeRefreshToken(ctx context.Context, refreshToken string) error {
	key := s.cfg.RefreshPrefix + refreshToken
	return s.rdb.Del(ctx, key).Err()
}

func (s *AuthService) parseAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
