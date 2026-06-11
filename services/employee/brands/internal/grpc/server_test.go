package grpc

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/dealer/dealer/services/brands/internal/domain"
	"github.com/dealer/dealer/services/brands/internal/service"
	brandsv1 "github.com/dealer/dealer/pkg/pb/brands/v1"
)

type mockGRPCBrand struct{}

func (mockGRPCBrand) Create(_ context.Context, name string) (*domain.Brand, error) {
	now := time.Now().UTC()
	return &domain.Brand{ID: uuid.New(), Name: name, CreatedAt: now, UpdatedAt: now}, nil
}

func (mockGRPCBrand) Get(_ context.Context, id string) (*domain.Brand, error) {
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	return &domain.Brand{ID: uid, Name: "B", CreatedAt: now, UpdatedAt: now}, nil
}

func (mockGRPCBrand) List(context.Context, int32, int32, string) ([]*domain.Brand, int32, error) {
	return nil, 0, nil
}

func (mockGRPCBrand) Update(_ context.Context, id string, name *string) (*domain.Brand, error) {
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	b := &domain.Brand{ID: uid, Name: "o", CreatedAt: now, UpdatedAt: now}
	if name != nil {
		b.Name = *name
	}
	return b, nil
}

func (mockGRPCBrand) Delete(context.Context, string) error { return nil }

func dial(t *testing.T, srv *grpc.Server) brandsv1.BrandsServiceClient {
	t.Helper()
	l := bufconn.Listen(1024 * 1024)
	go func() { _ = srv.Serve(l) }()
	t.Cleanup(func() { srv.Stop() })
	c, err := grpc.DialContext(context.Background(), "x",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return l.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return brandsv1.NewBrandsServiceClient(c)
}

func TestBrandsGRPC(t *testing.T) {
	s := grpc.NewServer()
	brandsv1.RegisterBrandsServiceServer(s, NewServer(mockGRPCBrand{}))
	cli := dial(t, s)
	ctx := context.Background()
	if _, err := cli.CreateBrand(ctx, &brandsv1.CreateBrandRequest{Name: "A"}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.GetBrand(ctx, &brandsv1.GetBrandRequest{Id: uuid.New().String()}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.ListBrands(ctx, &brandsv1.ListBrandsRequest{}); err != nil {
		t.Fatal(err)
	}
	n := "z"
	if _, err := cli.UpdateBrand(ctx, &brandsv1.UpdateBrandRequest{Id: uuid.New().String(), Name: &n}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.DeleteBrand(ctx, &brandsv1.DeleteBrandRequest{Id: uuid.New().String()}); err != nil {
		t.Fatal(err)
	}
}

func TestBrandsGRPC_GetNF(t *testing.T) {
	s := grpc.NewServer()
	brandsv1.RegisterBrandsServiceServer(s, NewServer(&stubBrandNF{}))
	cli := dial(t, s)
	_, err := cli.GetBrand(context.Background(), &brandsv1.GetBrandRequest{Id: uuid.New().String()})
	if err == nil {
		t.Fatal("want err")
	}
}

type stubBrandNF struct{ mockGRPCBrand }

func (stubBrandNF) Get(context.Context, string) (*domain.Brand, error) {
	return nil, service.ErrNotFound
}

func TestBrandsGRPC_UpdateNF(t *testing.T) {
	s := grpc.NewServer()
	brandsv1.RegisterBrandsServiceServer(s, NewServer(&stubBrandUpdateNF{}))
	cli := dial(t, s)
	n := "x"
	_, err := cli.UpdateBrand(context.Background(), &brandsv1.UpdateBrandRequest{Id: uuid.New().String(), Name: &n})
	if err == nil || status.Code(err) != codes.NotFound {
		t.Fatalf("%v", err)
	}
}

type stubBrandUpdateNF struct{ mockGRPCBrand }

func (stubBrandUpdateNF) Update(context.Context, string, *string) (*domain.Brand, error) {
	return nil, service.ErrNotFound
}

func TestBrandsGRPC_DeleteNF(t *testing.T) {
	s := grpc.NewServer()
	brandsv1.RegisterBrandsServiceServer(s, NewServer(&stubBrandDeleteNF{}))
	cli := dial(t, s)
	_, err := cli.DeleteBrand(context.Background(), &brandsv1.DeleteBrandRequest{Id: uuid.New().String()})
	if err == nil || status.Code(err) != codes.NotFound {
		t.Fatalf("%v", err)
	}
}

type stubBrandDeleteNF struct{ mockGRPCBrand }

func (stubBrandDeleteNF) Delete(context.Context, string) error {
	return service.ErrNotFound
}

func TestBrandsGRPC_ListErr(t *testing.T) {
	s := grpc.NewServer()
	brandsv1.RegisterBrandsServiceServer(s, NewServer(&stubBrandListErr{}))
	cli := dial(t, s)
	_, err := cli.ListBrands(context.Background(), &brandsv1.ListBrandsRequest{})
	if err == nil || status.Code(err) != codes.Internal {
		t.Fatalf("%v", err)
	}
}

type stubBrandListErr struct{ mockGRPCBrand }

func (stubBrandListErr) List(context.Context, int32, int32, string) ([]*domain.Brand, int32, error) {
	return nil, 0, errStubDB
}

var errStubDB = errors.New("db")

func TestBrandsGRPC_CreateErr(t *testing.T) {
	s := grpc.NewServer()
	brandsv1.RegisterBrandsServiceServer(s, NewServer(&stubBrandCreateErr{}))
	cli := dial(t, s)
	_, err := cli.CreateBrand(context.Background(), &brandsv1.CreateBrandRequest{Name: "A"})
	if err == nil || status.Code(err) != codes.Internal {
		t.Fatalf("%v", err)
	}
}

type stubBrandCreateErr struct{ mockGRPCBrand }

func (stubBrandCreateErr) Create(context.Context, string) (*domain.Brand, error) {
	return nil, errStubDB
}
