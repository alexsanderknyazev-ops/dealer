package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/dealer/dealer/auth-service/internal/domain"
)

type fakeUserRepo struct {
	byEmail map[string]*domain.User
	byID    map[uuid.UUID]*domain.User
	err     error
}

func (f *fakeUserRepo) Create(_ context.Context, u *domain.User) error {
	if f.err != nil {
		return f.err
	}
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

func (f *fakeUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	u, ok := f.byEmail[email]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return u, nil
}

func (f *fakeUserRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	u, ok := f.byID[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return u, nil
}

type pubSpy struct{ n int }

func (p *pubSpy) Publish(context.Context, []byte, []byte) error {
	p.n++
	return nil
}

func testRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { mr.Close() })
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return mr, rdb
}

func testCfg() AuthConfig {
	return AuthConfig{
		JWTSecret:     "secret",
		AccessTTL:     time.Hour,
		RefreshTTL:    time.Hour,
		RefreshPrefix: "rt:",
	}
}

func TestRegister_Login_Refresh_Logout(t *testing.T) {
	_, rdb := testRedis(t)
	repo := &fakeUserRepo{byEmail: map[string]*domain.User{}, byID: map[uuid.UUID]*domain.User{}}
	pub := &pubSpy{}
	s := NewAuthService(repo, rdb, pub, testCfg())
	ctx := context.Background()

	u, at, rt, exp, err := s.Register(ctx, "u@test.local", "pass12345", "U", "")
	if err != nil || at == "" || rt == "" || exp.IsZero() || u.Email != "u@test.local" {
		t.Fatalf("reg err=%v at=%q", err, at)
	}
	if pub.n != 1 {
		t.Fatalf("publish calls %d", pub.n)
	}

	_, _, _, _, err = s.Register(ctx, "u@test.local", "x", "x", "")
	if err != ErrUserExists {
		t.Fatalf("want exists got %v", err)
	}

	u2, at2, rt2, _, err := s.Login(ctx, "u@test.local", "pass12345")
	if err != nil || u2.ID != u.ID || at2 == "" || rt2 == "" {
		t.Fatalf("login %v", err)
	}
	_, _, _, _, err = s.Login(ctx, "u@test.local", "wrong")
	if err != ErrBadCredentials {
		t.Fatalf("login bad %v", err)
	}

	at3, rt3, _, err := s.Refresh(ctx, rt2)
	if err != nil || at3 == "" || rt3 == "" {
		t.Fatalf("refresh %v", err)
	}

	uid, email, ok := s.Validate(ctx, at3)
	if !ok || uid != u.ID.String() || email != u.Email {
		t.Fatalf("validate %v %q %q", ok, uid, email)
	}

	if err := s.Logout(ctx, rt3); err != nil {
		t.Fatal(err)
	}
}

func TestRegister_GetByEmail_Err(t *testing.T) {
	_, rdb := testRedis(t)
	s := NewAuthService(&fakeUserRepo{err: errors.New("db")}, rdb, nil, testCfg())
	_, _, _, _, err := s.Register(context.Background(), "a@b.c", "p", "n", "")
	if err == nil {
		t.Fatal("want err")
	}
}

func TestLogin_NoUser(t *testing.T) {
	_, rdb := testRedis(t)
	s := NewAuthService(&fakeUserRepo{byEmail: map[string]*domain.User{}}, rdb, nil, testCfg())
	_, _, _, _, err := s.Login(context.Background(), "missing@test", "x")
	if err != ErrBadCredentials {
		t.Fatalf("%v", err)
	}
}

func TestLogin_BadPassword(t *testing.T) {
	_, rdb := testRedis(t)
	hash, _ := bcrypt.GenerateFromPassword([]byte("good"), bcrypt.DefaultCost)
	id := uuid.New()
	now := time.Now().UTC()
	u := &domain.User{
		ID: id, Email: "e@e", PasswordHash: string(hash), Name: "n", Role: "sales", CreatedAt: now, UpdatedAt: now,
	}
	repo := &fakeUserRepo{byEmail: map[string]*domain.User{"e@e": u}, byID: map[uuid.UUID]*domain.User{id: u}}
	s := NewAuthService(repo, rdb, nil, testCfg())
	_, _, _, _, err := s.Login(context.Background(), "e@e", "bad")
	if err != ErrBadCredentials {
		t.Fatalf("%v", err)
	}
}

func TestRefresh_Invalid(t *testing.T) {
	_, rdb := testRedis(t)
	s := NewAuthService(&fakeUserRepo{}, rdb, nil, testCfg())
	_, _, _, err := s.Refresh(context.Background(), "nope")
	if err != ErrInvalidToken {
		t.Fatalf("%v", err)
	}
}

func TestValidate_BadToken(t *testing.T) {
	_, rdb := testRedis(t)
	s := NewAuthService(&fakeUserRepo{}, rdb, nil, testCfg())
	_, _, ok := s.Validate(context.Background(), "garbage")
	if ok {
		t.Fatal("want invalid")
	}
}
