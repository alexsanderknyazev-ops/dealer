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

	"github.com/dealer/dealer/services/vehicles/internal/domain"
	"github.com/dealer/dealer/services/vehicles/internal/service"
	vehiclesv1 "github.com/dealer/dealer/pkg/pb/vehicles/v1"
)

type mockGRPCVeh struct{}

func (mockGRPCVeh) Create(_ context.Context, in service.CreateVehicleInput) (*domain.Vehicle, error) {
	now := time.Now().UTC()
	return &domain.Vehicle{ID: uuid.New(), VIN: in.VIN, Make: in.Make, Model: in.Model, Year: in.Year, MileageKm: in.MileageKm, Price: in.Price, Status: in.Status, Color: in.Color, Notes: in.Notes, BrandID: in.BrandID, DealerPointID: in.DealerPointID, LegalEntityID: in.LegalEntityID, WarehouseID: in.WarehouseID, CreatedAt: now, UpdatedAt: now}, nil
}

func (mockGRPCVeh) Get(_ context.Context, id string) (*domain.Vehicle, error) {
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	return &domain.Vehicle{ID: uid, VIN: "v", Make: "m", Model: "m", Year: 1, Status: "a", CreatedAt: now, UpdatedAt: now}, nil
}

func (mockGRPCVeh) List(context.Context, domain.VehicleListFilter) ([]*domain.Vehicle, int32, error) {
	return nil, 0, nil
}

func (mockGRPCVeh) Update(_ context.Context, id string, _ service.UpdateVehicleInput) (*domain.Vehicle, error) {
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	return &domain.Vehicle{ID: uid, VIN: "v", Make: "m", Model: "m", Year: 1, Status: "a", CreatedAt: now, UpdatedAt: now}, nil
}

func (mockGRPCVeh) Delete(context.Context, string) error { return nil }

func dialVeh(t *testing.T, srv *grpc.Server) vehiclesv1.VehiclesServiceClient {
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
	return vehiclesv1.NewVehiclesServiceClient(c)
}

func TestVehiclesGRPC(t *testing.T) {
	s := grpc.NewServer()
	vehiclesv1.RegisterVehiclesServiceServer(s, NewServer(mockGRPCVeh{}))
	cli := dialVeh(t, s)
	ctx := context.Background()
	if _, err := cli.CreateVehicle(ctx, &vehiclesv1.CreateVehicleRequest{Vin: "V", Make: "M", Model: "X", Year: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.GetVehicle(ctx, &vehiclesv1.GetVehicleRequest{Id: uuid.New().String()}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.ListVehicles(ctx, &vehiclesv1.ListVehiclesRequest{}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.UpdateVehicle(ctx, &vehiclesv1.UpdateVehicleRequest{Id: uuid.New().String()}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.DeleteVehicle(ctx, &vehiclesv1.DeleteVehicleRequest{Id: uuid.New().String()}); err != nil {
		t.Fatal(err)
	}
}

func TestVehiclesGRPC_GetNF(t *testing.T) {
	s := grpc.NewServer()
	vehiclesv1.RegisterVehiclesServiceServer(s, NewServer(&stubVehNF{}))
	cli := dialVeh(t, s)
	_, err := cli.GetVehicle(context.Background(), &vehiclesv1.GetVehicleRequest{Id: uuid.New().String()})
	if err == nil || status.Code(err) != codes.NotFound {
		t.Fatalf("got %v", err)
	}
}

type stubVehNF struct{ mockGRPCVeh }

func (stubVehNF) Get(context.Context, string) (*domain.Vehicle, error) {
	return nil, service.ErrNotFound
}

func TestVehiclesGRPC_CreateErr(t *testing.T) {
	s := grpc.NewServer()
	vehiclesv1.RegisterVehiclesServiceServer(s, NewServer(&stubVehCreateErr{}))
	cli := dialVeh(t, s)
	_, err := cli.CreateVehicle(context.Background(), &vehiclesv1.CreateVehicleRequest{Vin: "x"})
	if err == nil || status.Code(err) != codes.Internal {
		t.Fatalf("%v", err)
	}
}

type stubVehCreateErr struct{ mockGRPCVeh }

func (stubVehCreateErr) Create(context.Context, service.CreateVehicleInput) (*domain.Vehicle, error) {
	return nil, errors.New("db")
}

func TestVehiclesGRPC_GetInternal(t *testing.T) {
	s := grpc.NewServer()
	vehiclesv1.RegisterVehiclesServiceServer(s, NewServer(&stubVehGetInt{}))
	cli := dialVeh(t, s)
	_, err := cli.GetVehicle(context.Background(), &vehiclesv1.GetVehicleRequest{Id: uuid.New().String()})
	if err == nil || status.Code(err) != codes.Internal {
		t.Fatalf("%v", err)
	}
}

type stubVehGetInt struct{ mockGRPCVeh }

func (stubVehGetInt) Get(context.Context, string) (*domain.Vehicle, error) {
	return nil, errors.New("db")
}

func TestVehiclesGRPC_ListErr(t *testing.T) {
	s := grpc.NewServer()
	vehiclesv1.RegisterVehiclesServiceServer(s, NewServer(&stubVehListErr{}))
	cli := dialVeh(t, s)
	_, err := cli.ListVehicles(context.Background(), &vehiclesv1.ListVehiclesRequest{})
	if err == nil || status.Code(err) != codes.Internal {
		t.Fatalf("%v", err)
	}
}

type stubVehListErr struct{ mockGRPCVeh }

func (stubVehListErr) List(context.Context, domain.VehicleListFilter) ([]*domain.Vehicle, int32, error) {
	return nil, 0, errors.New("db")
}

func TestVehiclesGRPC_UpdateNF(t *testing.T) {
	s := grpc.NewServer()
	vehiclesv1.RegisterVehiclesServiceServer(s, NewServer(&stubVehUpdateNF{}))
	cli := dialVeh(t, s)
	mk := "z"
	_, err := cli.UpdateVehicle(context.Background(), &vehiclesv1.UpdateVehicleRequest{Id: uuid.New().String(), Make: &mk})
	if err == nil || status.Code(err) != codes.NotFound {
		t.Fatalf("%v", err)
	}
}

type stubVehUpdateNF struct{ mockGRPCVeh }

func (stubVehUpdateNF) Update(context.Context, string, service.UpdateVehicleInput) (*domain.Vehicle, error) {
	return nil, service.ErrNotFound
}

func TestVehiclesGRPC_UpdateWithClears(t *testing.T) {
	s := grpc.NewServer()
	vehiclesv1.RegisterVehiclesServiceServer(s, NewServer(mockGRPCVeh{}))
	cli := dialVeh(t, s)
	empty := ""
	bid := uuid.New().String()
	_, err := cli.UpdateVehicle(context.Background(), &vehiclesv1.UpdateVehicleRequest{
		Id: uuid.New().String(), BrandId: &empty, DealerPointId: &empty, LegalEntityId: &empty, WarehouseId: &empty,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cli.UpdateVehicle(context.Background(), &vehiclesv1.UpdateVehicleRequest{
		Id: uuid.New().String(), BrandId: &bid,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestVehiclesGRPC_DeleteNF(t *testing.T) {
	s := grpc.NewServer()
	vehiclesv1.RegisterVehiclesServiceServer(s, NewServer(&stubVehDeleteNF{}))
	cli := dialVeh(t, s)
	_, err := cli.DeleteVehicle(context.Background(), &vehiclesv1.DeleteVehicleRequest{Id: uuid.New().String()})
	if err == nil || status.Code(err) != codes.NotFound {
		t.Fatalf("%v", err)
	}
}

type stubVehDeleteNF struct{ mockGRPCVeh }

func (stubVehDeleteNF) Delete(context.Context, string) error {
	return service.ErrNotFound
}

func TestVehiclesGRPC_DeleteInternal(t *testing.T) {
	s := grpc.NewServer()
	vehiclesv1.RegisterVehiclesServiceServer(s, NewServer(&stubVehDeleteInt{}))
	cli := dialVeh(t, s)
	_, err := cli.DeleteVehicle(context.Background(), &vehiclesv1.DeleteVehicleRequest{Id: uuid.New().String()})
	if err == nil || status.Code(err) != codes.Internal {
		t.Fatalf("%v", err)
	}
}

type stubVehDeleteInt struct{ mockGRPCVeh }

func (stubVehDeleteInt) Delete(context.Context, string) error {
	return errors.New("db")
}

func TestVehiclesGRPC_toProtoOptionalIDs(t *testing.T) {
	s := grpc.NewServer()
	vehiclesv1.RegisterVehiclesServiceServer(s, NewServer(&stubVehProto{}))
	cli := dialVeh(t, s)
	resp, err := cli.GetVehicle(context.Background(), &vehiclesv1.GetVehicleRequest{Id: uuid.New().String()})
	if err != nil || resp.GetVehicle().GetBrandId() == "" {
		t.Fatalf("%v", err)
	}
}

type stubVehProto struct{ mockGRPCVeh }

func (stubVehProto) Get(_ context.Context, id string) (*domain.Vehicle, error) {
	uid, _ := uuid.Parse(id)
	b, d, l, w := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	return &domain.Vehicle{
		ID: uid, VIN: "v", Make: "m", Model: "m", Year: 1, Status: "a",
		BrandID: &b, DealerPointID: &d, LegalEntityID: &l, WarehouseID: &w,
		CreatedAt: now, UpdatedAt: now,
	}, nil
}
