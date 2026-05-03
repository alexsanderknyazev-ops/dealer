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

	dealsv1 "github.com/dealer/dealer/pkg/pb/deals/v1"
	"github.com/dealer/dealer/services/deals/internal/domain"
	"github.com/dealer/dealer/services/deals/internal/service"
)

const (
	testDealAmountStub = "1"
	testDealStageStub  = "d"
)

type mockGRPCDeal struct{}

func (mockGRPCDeal) Create(_ context.Context, in service.CreateDealInput) (*domain.Deal, error) {
	now := time.Now().UTC()
	cid, _ := uuid.Parse(in.CustomerID)
	vid, _ := uuid.Parse(in.VehicleID)
	return &domain.Deal{ID: uuid.New(), CustomerID: cid, VehicleID: vid, Amount: in.Amount, Stage: in.Stage, Notes: in.Notes, CreatedAt: now, UpdatedAt: now}, nil
}

func (mockGRPCDeal) Get(_ context.Context, id string) (*domain.Deal, error) {
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	return &domain.Deal{ID: uid, CustomerID: uuid.New(), VehicleID: uuid.New(), Amount: testDealAmountStub, Stage: testDealStageStub, CreatedAt: now, UpdatedAt: now}, nil
}

func (mockGRPCDeal) List(context.Context, int32, int32, string, string) ([]*domain.Deal, int32, error) {
	return nil, 0, nil
}

func (mockGRPCDeal) Update(_ context.Context, id string, _ service.UpdateDealInput) (*domain.Deal, error) {
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	return &domain.Deal{ID: uid, CustomerID: uuid.New(), VehicleID: uuid.New(), Amount: testDealAmountStub, Stage: testDealStageStub, CreatedAt: now, UpdatedAt: now}, nil
}

func (mockGRPCDeal) Delete(context.Context, string) error { return nil }

func dialDeal(t *testing.T, srv *grpc.Server) dealsv1.DealsServiceClient {
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
	return dealsv1.NewDealsServiceClient(c)
}

func TestDealsGRPC(t *testing.T) {
	s := grpc.NewServer()
	dealsv1.RegisterDealsServiceServer(s, NewServer(mockGRPCDeal{}))
	cli := dialDeal(t, s)
	ctx := context.Background()
	cid, vid := uuid.New().String(), uuid.New().String()
	if _, err := cli.CreateDeal(ctx, &dealsv1.CreateDealRequest{CustomerId: cid, VehicleId: vid}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.GetDeal(ctx, &dealsv1.GetDealRequest{Id: uuid.New().String()}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.ListDeals(ctx, &dealsv1.ListDealsRequest{}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.UpdateDeal(ctx, &dealsv1.UpdateDealRequest{Id: uuid.New().String()}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.DeleteDeal(ctx, &dealsv1.DeleteDealRequest{Id: uuid.New().String()}); err != nil {
		t.Fatal(err)
	}
}

func TestDealsGRPC_GetNF(t *testing.T) {
	s := grpc.NewServer()
	dealsv1.RegisterDealsServiceServer(s, NewServer(&stubDealNF{}))
	cli := dialDeal(t, s)
	_, err := cli.GetDeal(context.Background(), &dealsv1.GetDealRequest{Id: uuid.New().String()})
	if err == nil || status.Code(err) != codes.NotFound {
		t.Fatalf("%v", err)
	}
}

type stubDealNF struct{ mockGRPCDeal }

func (stubDealNF) Get(context.Context, string) (*domain.Deal, error) {
	return nil, service.ErrNotFound
}

func TestDealsGRPC_CreateErr(t *testing.T) {
	s := grpc.NewServer()
	dealsv1.RegisterDealsServiceServer(s, NewServer(&stubDealCreateErr{}))
	cli := dialDeal(t, s)
	cid, vid := uuid.New().String(), uuid.New().String()
	_, err := cli.CreateDeal(context.Background(), &dealsv1.CreateDealRequest{CustomerId: cid, VehicleId: vid})
	if err == nil || status.Code(err) != codes.Internal {
		t.Fatalf("%v", err)
	}
}

type stubDealCreateErr struct{ mockGRPCDeal }

func (stubDealCreateErr) Create(context.Context, service.CreateDealInput) (*domain.Deal, error) {
	return nil, errors.New("db")
}

func TestDealsGRPC_GetInternal(t *testing.T) {
	s := grpc.NewServer()
	dealsv1.RegisterDealsServiceServer(s, NewServer(&stubDealGetInt{}))
	cli := dialDeal(t, s)
	_, err := cli.GetDeal(context.Background(), &dealsv1.GetDealRequest{Id: uuid.New().String()})
	if err == nil || status.Code(err) != codes.Internal {
		t.Fatalf("%v", err)
	}
}

type stubDealGetInt struct{ mockGRPCDeal }

func (stubDealGetInt) Get(context.Context, string) (*domain.Deal, error) {
	return nil, errors.New("db")
}

func TestDealsGRPC_ListErr(t *testing.T) {
	s := grpc.NewServer()
	dealsv1.RegisterDealsServiceServer(s, NewServer(&stubDealListErr{}))
	cli := dialDeal(t, s)
	_, err := cli.ListDeals(context.Background(), &dealsv1.ListDealsRequest{})
	if err == nil || status.Code(err) != codes.Internal {
		t.Fatalf("%v", err)
	}
}

type stubDealListErr struct{ mockGRPCDeal }

func (stubDealListErr) List(context.Context, int32, int32, string, string) ([]*domain.Deal, int32, error) {
	return nil, 0, errors.New("db")
}

func TestDealsGRPC_UpdateNF(t *testing.T) {
	s := grpc.NewServer()
	dealsv1.RegisterDealsServiceServer(s, NewServer(&stubDealUpdateNF{}))
	cli := dialDeal(t, s)
	a := testDealAmountStub
	_, err := cli.UpdateDeal(context.Background(), &dealsv1.UpdateDealRequest{Id: uuid.New().String(), Amount: &a})
	if err == nil || status.Code(err) != codes.NotFound {
		t.Fatalf("%v", err)
	}
}

type stubDealUpdateNF struct{ mockGRPCDeal }

func (stubDealUpdateNF) Update(context.Context, string, service.UpdateDealInput) (*domain.Deal, error) {
	return nil, service.ErrNotFound
}

func TestDealsGRPC_DeleteNF(t *testing.T) {
	s := grpc.NewServer()
	dealsv1.RegisterDealsServiceServer(s, NewServer(&stubDealDeleteNF{}))
	cli := dialDeal(t, s)
	_, err := cli.DeleteDeal(context.Background(), &dealsv1.DeleteDealRequest{Id: uuid.New().String()})
	if err == nil || status.Code(err) != codes.NotFound {
		t.Fatalf("%v", err)
	}
}

type stubDealDeleteNF struct{ mockGRPCDeal }

func (stubDealDeleteNF) Delete(context.Context, string) error {
	return service.ErrNotFound
}
