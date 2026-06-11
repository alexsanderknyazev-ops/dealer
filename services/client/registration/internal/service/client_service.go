package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dealer/dealer/client-registration-service/internal/domain"
	"github.com/dealer/dealer/client-registration-service/internal/repository"
	"github.com/dealer/dealer/pkg/clientevent"
	"github.com/dealer/dealer/pkg/obslog"
	vehiclesv1 "github.com/dealer/dealer/pkg/pb/vehicles/v1"
)

var (
	ErrVINNotFound      = errors.New("vehicle with this vin not found")
	ErrVINAlreadyLinked = repository.ErrVINAlreadyLinked
	ErrClientNotFound   = errors.New("client not found")
	ErrUserExists       = errors.New("user with this email already exists")
	ErrAuthNotReady     = errors.New("client auth not ready yet")
)

type registrationPublisher interface {
	Publish(ctx context.Context, key, value []byte) error
}

type sessionIssuer interface {
	IssueTokens(ctx context.Context, userID string) (access, refresh string, expiresAt int64, err error)
}

type vehicleLookup interface {
	GetByVIN(ctx context.Context, vin string) (*vehiclesv1.Vehicle, error)
}

type clientRepository interface {
	EmailExists(ctx context.Context, email string) (bool, error)
	CreateClientWithVehicle(ctx context.Context, c *domain.Client, v *domain.ClientVehicle) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Client, error)
	ListVehiclesByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.ClientVehicle, error)
	AddVehicle(ctx context.Context, clientID uuid.UUID, v *domain.ClientVehicle) error
	VINLinked(ctx context.Context, vin string) (bool, error)
}

type RegisterResult struct {
	Client       *domain.Client
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

type ClientAPI interface {
	Register(ctx context.Context, email, password, fullName, phone, vin string) (*RegisterResult, error)
	AddVehicle(ctx context.Context, userID uuid.UUID, vin string) (*domain.ClientVehicle, error)
	ListVehicles(ctx context.Context, userID uuid.UUID) ([]*domain.ClientVehicle, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.Client, []*domain.ClientVehicle, error)
}

type ClientService struct {
	repo       clientRepository
	publisher  registrationPublisher
	clientAuth sessionIssuer
	vehicles   vehicleLookup
}

func NewClientService(repo clientRepository, publisher registrationPublisher, clientAuth sessionIssuer, vehicles vehicleLookup) *ClientService {
	return &ClientService{repo: repo, publisher: publisher, clientAuth: clientAuth, vehicles: vehicles}
}

func (s *ClientService) Register(ctx context.Context, email, password, fullName, phone, vin string) (*RegisterResult, error) {
	vin = normalizeVIN(vin)
	veh, err := s.lookupVehicle(ctx, vin)
	if err != nil {
		return nil, err
	}
	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}
	linked, err := s.repo.VINLinked(ctx, vin)
	if err != nil {
		return nil, err
	}
	if linked {
		return nil, ErrVINAlreadyLinked
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID := uuid.New()
	vehicleID, err := uuid.Parse(veh.Id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	client := &domain.Client{
		ID:        uuid.New(),
		UserID:    userID,
		Email:     email,
		FullName:  fullName,
		Phone:     phone,
		CreatedAt: now,
		UpdatedAt: now,
	}
	cv := &domain.ClientVehicle{
		ID:        uuid.New(),
		VehicleID: vehicleID,
		VIN:       vin,
		Make:      veh.Make,
		Model:     veh.Model,
		Year:      veh.Year,
		AddedAt:   now,
	}
	if err := s.repo.CreateClientWithVehicle(ctx, client, cv); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrUserExists
		}
		return nil, err
	}

	if err := s.publishRegistered(ctx, userID, email, string(hash), fullName, phone, vehicleID.String()); err != nil {
		obslog.Default.Warn("kafka publish failed", "event", clientevent.Registered, "user_id", userID.String(), "err", err)
		return nil, err
	}

	access, refresh, expiresAt, err := issueTokensWithRetry(ctx, s.clientAuth, userID.String())
	if err != nil {
		return nil, ErrAuthNotReady
	}

	return &RegisterResult{
		Client:       client,
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *ClientService) publishRegistered(ctx context.Context, userID uuid.UUID, email, passwordHash, fullName, phone, vehicleID string) error {
	if s.publisher == nil {
		return errors.New("registration publisher not configured")
	}
	ev := clientevent.RegisteredEvent{
		Event:        clientevent.Registered,
		UserID:       userID.String(),
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Phone:        phone,
		VehicleID:    vehicleID,
	}
	body, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	return s.publisher.Publish(ctx, []byte(userID.String()), body)
}

func issueTokensWithRetry(ctx context.Context, issuer sessionIssuer, userID string) (access, refresh string, expiresAt int64, err error) {
	const attempts = 15
	for i := 0; i < attempts; i++ {
		access, refresh, expiresAt, err = issuer.IssueTokens(ctx, userID)
		if err == nil {
			return access, refresh, expiresAt, nil
		}
		select {
		case <-ctx.Done():
			return "", "", 0, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
	return "", "", 0, err
}

func (s *ClientService) AddVehicle(ctx context.Context, userID uuid.UUID, vin string) (*domain.ClientVehicle, error) {
	vin = normalizeVIN(vin)
	client, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrClientNotFound
		}
		return nil, err
	}

	veh, err := s.lookupVehicle(ctx, vin)
	if err != nil {
		return nil, err
	}
	linked, err := s.repo.VINLinked(ctx, vin)
	if err != nil {
		return nil, err
	}
	if linked {
		return nil, ErrVINAlreadyLinked
	}

	vehicleID, err := uuid.Parse(veh.Id)
	if err != nil {
		return nil, err
	}
	cv := &domain.ClientVehicle{
		VehicleID: vehicleID,
		VIN:       vin,
		Make:      veh.Make,
		Model:     veh.Model,
		Year:      veh.Year,
		AddedAt:   time.Now().UTC(),
	}
	if err := s.repo.AddVehicle(ctx, client.ID, cv); err != nil {
		return nil, err
	}
	return cv, nil
}

func (s *ClientService) ListVehicles(ctx context.Context, userID uuid.UUID) ([]*domain.ClientVehicle, error) {
	client, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrClientNotFound
		}
		return nil, err
	}
	vehicles, err := s.repo.ListVehiclesByClientID(ctx, client.ID)
	if err != nil {
		return nil, err
	}
	return vehicles, nil
}

func (s *ClientService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.Client, []*domain.ClientVehicle, error) {
	client, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrClientNotFound
		}
		return nil, nil, err
	}
	vehicles, err := s.repo.ListVehiclesByClientID(ctx, client.ID)
	if err != nil {
		return nil, nil, err
	}
	return client, vehicles, nil
}

func (s *ClientService) lookupVehicle(ctx context.Context, vin string) (*vehiclesv1.Vehicle, error) {
	veh, err := s.vehicles.GetByVIN(ctx, vin)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrVINNotFound
		}
		return nil, err
	}
	return veh, nil
}

func normalizeVIN(vin string) string {
	return strings.ToUpper(strings.TrimSpace(vin))
}

func isUniqueViolation(err error) bool {
	var pgErr interface{ SQLState() string }
	return errors.As(err, &pgErr) && pgErr.SQLState() == "23505"
}
