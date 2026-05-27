package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/dealer/dealer/auth-service/internal/domain"
	"github.com/dealer/dealer/auth-service/internal/service"
	authv1 "github.com/dealer/dealer/pkg/pb/auth/v1"
)

const (
	testGRPCJWTSecret       = "s"
	testGRPCRefreshPrefix   = "rt:"
	testGRPCRegisterPass    = "password123"
	testGRPCEmailOK       = "g@grpc.test"
	testGRPCEmailDup      = "dup@x"
)

type grpcUserFake struct {
	byEmail map[string]*domain.User
	byID    map[uuid.UUID]*domain.User
}

func (f *grpcUserFake) Create(_ context.Context, u *domain.User) error {
	if f.byEmail == nil {
		f.byEmail = make(map[string]*domain.User)
	}
	if f.byID == nil {
		f.byID = make(map[uuid.UUID]*domain.User)
	}
	cp := *u
	f.byEmail[u.Email] = &cp
	f.byID[u.ID] = &cp
	return nil
}

func (f *grpcUserFake) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return u, nil
}

func (f *grpcUserFake) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return u, nil
}

func dialAuth(t *testing.T, srv *grpc.Server) authv1.AuthServiceClient {
	t.Helper()
	l := bufconn.Listen(1024 * 1024)
	go func() { _ = srv.Serve(l) }()
	t.Cleanup(func() { srv.Stop() })
	c, err := grpc.DialContext(context.Background(), "buf",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return l.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return authv1.NewAuthServiceClient(c)
}

func newAuthSvc(t *testing.T) (*service.AuthService, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := &grpcUserFake{byEmail: map[string]*domain.User{}, byID: map[uuid.UUID]*domain.User{}}
	svc := service.NewAuthService(repo, rdb, nil, service.AuthConfig{
		JWTSecret: testGRPCJWTSecret, AccessTTL: time.Hour, RefreshTTL: time.Hour, RefreshPrefix: testGRPCRefreshPrefix,
	})
	cleanup := func() {
		_ = rdb.Close()
		mr.Close()
	}
	return svc, cleanup
}

func TestAuthGRPC_RegisterValidation(t *testing.T) {
	svc, cleanup := newAuthSvc(t)
	defer cleanup()
	s := grpc.NewServer()
	authv1.RegisterAuthServiceServer(s, NewServer(svc))
	cli := dialAuth(t, s)
	ctx := context.Background()
	_, err := cli.Register(ctx, &authv1.RegisterRequest{Email: "", Password: "x"})
	if err == nil || status.Code(err) != codes.InvalidArgument {
		t.Fatalf("%v", err)
	}
}

func TestAuthGRPC_RegisterLoginRefreshValidate(t *testing.T) {
	svc, cleanup := newAuthSvc(t)
	defer cleanup()
	s := grpc.NewServer()
	authv1.RegisterAuthServiceServer(s, NewServer(svc))
	cli := dialAuth(t, s)
	ctx := context.Background()
	reg, err := cli.Register(ctx, &authv1.RegisterRequest{Email: testGRPCEmailOK, Password: testGRPCRegisterPass, Name: "G"})
	if err != nil {
		t.Fatal(err)
	}
	if reg.AccessToken == "" || reg.RefreshToken == "" {
		t.Fatal("tokens")
	}
	_, err = cli.Login(ctx, &authv1.LoginRequest{Email: testGRPCEmailOK, Password: "bad"})
	if err == nil || status.Code(err) != codes.Unauthenticated {
		t.Fatalf("login %v", err)
	}
	_, err = cli.Refresh(ctx, &authv1.RefreshRequest{})
	if err == nil || status.Code(err) != codes.InvalidArgument {
		t.Fatalf("refresh empty %v", err)
	}
	_, err = cli.Refresh(ctx, &authv1.RefreshRequest{RefreshToken: "nope"})
	if err == nil || status.Code(err) != codes.Unauthenticated {
		t.Fatalf("refresh bad %v", err)
	}
	ref, err := cli.Refresh(ctx, &authv1.RefreshRequest{RefreshToken: reg.RefreshToken})
	if err != nil {
		t.Fatal(err)
	}
	v, err := cli.Validate(ctx, &authv1.ValidateRequest{AccessToken: ""})
	if err != nil || v.Valid {
		t.Fatalf("%v %v", err, v)
	}
	v2, err := cli.Validate(ctx, &authv1.ValidateRequest{AccessToken: ref.AccessToken})
	if err != nil || !v2.Valid {
		t.Fatalf("%v", v2)
	}
	_, _ = cli.Logout(ctx, &authv1.LogoutRequest{RefreshToken: ref.RefreshToken})
}

func TestAuthGRPC_RegisterDuplicate(t *testing.T) {
	svc, cleanup := newAuthSvc(t)
	defer cleanup()
	s := grpc.NewServer()
	authv1.RegisterAuthServiceServer(s, NewServer(svc))
	cli := dialAuth(t, s)
	ctx := context.Background()
	_, err := cli.Register(ctx, &authv1.RegisterRequest{Email: testGRPCEmailDup, Password: testGRPCRegisterPass, Name: "A"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cli.Register(ctx, &authv1.RegisterRequest{Email: testGRPCEmailDup, Password: testGRPCRegisterPass, Name: "B"})
	if err == nil || status.Code(err) != codes.AlreadyExists {
		t.Fatalf("%v", err)
	}
}

func TestAuthGRPC_LoginValidation(t *testing.T) {
	svc, cleanup := newAuthSvc(t)
	defer cleanup()
	s := grpc.NewServer()
	authv1.RegisterAuthServiceServer(s, NewServer(svc))
	cli := dialAuth(t, s)
	_, err := cli.Login(context.Background(), &authv1.LoginRequest{Email: "", Password: ""})
	if err == nil || status.Code(err) != codes.InvalidArgument {
		t.Fatalf("%v", err)
	}
}
