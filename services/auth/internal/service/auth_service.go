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

	"github.com/dealer/dealer/auth-service/internal/domain"
)

var (
	ErrUserExists      = errors.New("user with this email already exists")
	ErrBadCredentials  = errors.New("invalid email or password")
	ErrInvalidToken    = errors.New("invalid or expired token")
)

// JWTClaims — поля в access JWT.
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// AuthConfig — настройки JWT и сессий.
type AuthConfig struct {
	JWTSecret     string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	RefreshPrefix string
}

// EventPublisher публикует события (например, в Kafka).
type EventPublisher interface {
	Publish(ctx context.Context, key, value []byte) error
}

type userRepository interface {
	Create(ctx context.Context, u *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

// AuthService — бизнес-логика авторизации и аутентификации.
type AuthService struct {
	repo      userRepository
	rdb       *redis.Client
	publisher EventPublisher
	cfg       AuthConfig
}

// NewAuthService создаёт сервис авторизации.
func NewAuthService(
	repo userRepository,
	rdb *redis.Client,
	publisher EventPublisher,
	cfg AuthConfig,
) *AuthService {
	return &AuthService{repo: repo, rdb: rdb, publisher: publisher, cfg: cfg}
}

// Register регистрирует пользователя, создаёт пару токенов и публикует событие в Kafka.
func (s *AuthService) Register(ctx context.Context, email, password, name, phone string) (*domain.User, string, string, time.Time, error) {
	_, err := s.repo.GetByEmail(ctx, email)
	if err == nil {
		return nil, "", "", time.Time{}, ErrUserExists
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, "", "", time.Time{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", time.Time{}, err
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Phone:        phone,
		Role:         domain.DefaultRole,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, "", "", time.Time{}, err
	}

	accessToken, expiresAt, err := s.issueAccessToken(user)
	if err != nil {
		return user, "", "", time.Time{}, err
	}
	refreshToken, err := s.issueRefreshToken(ctx, user)
	if err != nil {
		return user, "", "", time.Time{}, err
	}

	_ = s.publishUserRegistered(ctx, user)
	return user, accessToken, refreshToken, expiresAt, nil
}

// Login проверяет пароль и выдаёт пару токенов.
func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, string, string, time.Time, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", time.Time{}, ErrBadCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", time.Time{}, ErrBadCredentials
	}

	accessToken, expiresAt, err := s.issueAccessToken(user)
	if err != nil {
		return nil, "", "", time.Time{}, err
	}
	refreshToken, err := s.issueRefreshToken(ctx, user)
	if err != nil {
		return nil, "", "", time.Time{}, err
	}
	return user, accessToken, refreshToken, expiresAt, nil
}

// Refresh обновляет access по refresh-токену (проверяет Redis).
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

// Logout инвалидирует refresh-токен в Redis.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.revokeRefreshToken(ctx, refreshToken)
}

// Validate проверяет access-токен и возвращает user_id, email.
func (s *AuthService) Validate(ctx context.Context, accessToken string) (userID, email string, valid bool) {
	claims, err := s.parseAccessToken(accessToken)
	if err != nil {
		return "", "", false
	}
	return claims.UserID, claims.Email, true
}

func (s *AuthService) issueAccessToken(u *domain.User) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(s.cfg.AccessTTL)
	claims := &JWTClaims{
		UserID: u.ID.String(),
		Email:  u.Email,
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
	ttl := s.cfg.RefreshTTL
	payload := map[string]string{"user_id": u.ID.String(), "email": u.Email}
	data, _ := json.Marshal(payload)
	if err := s.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
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
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
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

func (s *AuthService) publishUserRegistered(ctx context.Context, u *domain.User) error {
	if s.publisher == nil {
		return nil
	}
	event := map[string]string{
		"event":   "user.registered",
		"user_id": u.ID.String(),
		"email":   u.Email,
	}
	body, _ := json.Marshal(event)
	return s.publisher.Publish(ctx, []byte(u.ID.String()), body)
}
