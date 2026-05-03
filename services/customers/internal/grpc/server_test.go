package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/dealer/dealer/customers-service/internal/domain"
	"github.com/dealer/dealer/customers-service/internal/service"
	customersv1 "github.com/dealer/dealer/pkg/pb/customers/v1"
)

const (
	testGRPCCustomerName  = "A"
	testGRPCCustomerEmail = "a@b.c"
	testWantErr           = "want err"
)

type grpcCustomerMock struct{}

func (grpcCustomerMock) Create(_ context.Context, in service.CreateCustomerInput) (*domain.Customer, error) {
	now := time.Now().UTC()
	return &domain.Customer{
		ID: uuid.New(), Name: in.Name, Email: in.Email, Phone: in.Phone, CustomerType: in.CustomerType,
		INN: in.INN, Address: in.Address, Notes: in.Notes, CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (grpcCustomerMock) Get(_ context.Context, id string) (*domain.Customer, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, service.ErrNotFound
	}
	now := time.Now().UTC()
	return &domain.Customer{ID: uid, Name: "N", Email: "e", CreatedAt: now, UpdatedAt: now}, nil
}

func (grpcCustomerMock) List(_ context.Context, _ domain.CustomerListParams) ([]*domain.Customer, int32, error) {
	return []*domain.Customer{}, 0, nil
}

func (grpcCustomerMock) Update(_ context.Context, id string, in service.UpdateCustomerInput) (*domain.Customer, error) {
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	c := &domain.Customer{ID: uid, Name: "old", Email: "e", CreatedAt: now, UpdatedAt: now}
	if in.Name != nil {
		c.Name = *in.Name
	}
	if in.Email != nil {
		c.Email = *in.Email
	}
	if in.Phone != nil {
		c.Phone = *in.Phone
	}
	if in.CustomerType != nil {
		c.CustomerType = *in.CustomerType
	}
	if in.INN != nil {
		c.INN = *in.INN
	}
	if in.Address != nil {
		c.Address = *in.Address
	}
	if in.Notes != nil {
		c.Notes = *in.Notes
	}
	return c, nil
}

func (grpcCustomerMock) Delete(_ context.Context, id string) error {
	return nil
}

func dialTestServer(t *testing.T, srv *grpc.Server) customersv1.CustomersServiceClient {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop() })
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return customersv1.NewCustomersServiceClient(conn)
}

func TestServer_CreateCustomer(t *testing.T) {
	s := grpc.NewServer()
	customersv1.RegisterCustomersServiceServer(s, NewServer(grpcCustomerMock{}))
	cli := dialTestServer(t, s)
	resp, err := cli.CreateCustomer(context.Background(), &customersv1.CreateCustomerRequest{
		Name: testGRPCCustomerName, Email: testGRPCCustomerEmail,
	})
	if err != nil || resp.GetCustomer().GetName() != testGRPCCustomerName {
		t.Fatalf("err=%v resp=%v", err, resp)
	}
}

func TestServer_GetCustomer_NotFound(t *testing.T) {
	s := grpc.NewServer()
	customersv1.RegisterCustomersServiceServer(s, NewServer(&stubNotFound{}))
	cli := dialTestServer(t, s)
	_, err := cli.GetCustomer(context.Background(), &customersv1.GetCustomerRequest{Id: uuid.New().String()})
	if err == nil {
		t.Fatal(testWantErr)
	}
}

type stubNotFound struct{ grpcCustomerMock }

func (stubNotFound) Get(_ context.Context, _ string) (*domain.Customer, error) {
	return nil, service.ErrNotFound
}

func TestServer_ListCustomers(t *testing.T) {
	s := grpc.NewServer()
	customersv1.RegisterCustomersServiceServer(s, NewServer(grpcCustomerMock{}))
	cli := dialTestServer(t, s)
	resp, err := cli.ListCustomers(context.Background(), &customersv1.ListCustomersRequest{Limit: 10, Offset: 0})
	if err != nil || resp.Total != 0 {
		t.Fatalf("err=%v total=%d", err, resp.Total)
	}
}

func TestServer_UpdateCustomer_NotFound(t *testing.T) {
	s := grpc.NewServer()
	customersv1.RegisterCustomersServiceServer(s, NewServer(&stubUpdateNF{}))
	cli := dialTestServer(t, s)
	n := "x"
	_, err := cli.UpdateCustomer(context.Background(), &customersv1.UpdateCustomerRequest{Id: uuid.New().String(), Name: &n})
	if err == nil {
		t.Fatal(testWantErr)
	}
}

type stubUpdateNF struct{ grpcCustomerMock }

func (stubUpdateNF) Update(context.Context, string, service.UpdateCustomerInput) (*domain.Customer, error) {
	return nil, service.ErrNotFound
}

func TestServer_UpdateCustomer_OK(t *testing.T) {
	s := grpc.NewServer()
	customersv1.RegisterCustomersServiceServer(s, NewServer(grpcCustomerMock{}))
	cli := dialTestServer(t, s)
	id := uuid.New().String()
	n := "new"
	resp, err := cli.UpdateCustomer(context.Background(), &customersv1.UpdateCustomerRequest{Id: id, Name: &n})
	if err != nil || resp.GetCustomer().GetName() != "new" {
		t.Fatalf("err=%v", err)
	}
}

func TestServer_DeleteCustomer_NotFound(t *testing.T) {
	s := grpc.NewServer()
	customersv1.RegisterCustomersServiceServer(s, NewServer(&stubDeleteNF{}))
	cli := dialTestServer(t, s)
	_, err := cli.DeleteCustomer(context.Background(), &customersv1.DeleteCustomerRequest{Id: uuid.New().String()})
	if err == nil {
		t.Fatal(testWantErr)
	}
}

func TestServer_DeleteCustomer_OK(t *testing.T) {
	s := grpc.NewServer()
	customersv1.RegisterCustomersServiceServer(s, NewServer(grpcCustomerMock{}))
	cli := dialTestServer(t, s)
	_, err := cli.DeleteCustomer(context.Background(), &customersv1.DeleteCustomerRequest{Id: uuid.New().String()})
	if err != nil {
		t.Fatal(err)
	}
}

type stubDeleteNF struct{ grpcCustomerMock }

func (stubDeleteNF) Delete(_ context.Context, _ string) error {
	return service.ErrNotFound
}

func TestServer_CreateCustomer_Err(t *testing.T) {
	s := grpc.NewServer()
	customersv1.RegisterCustomersServiceServer(s, NewServer(&stubCreateErr{}))
	cli := dialTestServer(t, s)
	_, err := cli.CreateCustomer(context.Background(), &customersv1.CreateCustomerRequest{Name: testGRPCCustomerName})
	if err == nil {
		t.Fatal(testWantErr)
	}
}

type stubCreateErr struct{ grpcCustomerMock }

func (stubCreateErr) Create(context.Context, service.CreateCustomerInput) (*domain.Customer, error) {
	return nil, context.Canceled
}

func TestServer_ListCustomers_InternalErr(t *testing.T) {
	s := grpc.NewServer()
	customersv1.RegisterCustomersServiceServer(s, NewServer(&stubListErr{}))
	cli := dialTestServer(t, s)
	_, err := cli.ListCustomers(context.Background(), &customersv1.ListCustomersRequest{})
	if err == nil {
		t.Fatal(testWantErr)
	}
}

type stubListErr struct{ grpcCustomerMock }

func (stubListErr) List(context.Context, domain.CustomerListParams) ([]*domain.Customer, int32, error) {
	return nil, 0, context.Canceled
}

func TestServer_GetCustomer_InternalErr(t *testing.T) {
	s := grpc.NewServer()
	customersv1.RegisterCustomersServiceServer(s, NewServer(&stubGetInt{}))
	cli := dialTestServer(t, s)
	_, err := cli.GetCustomer(context.Background(), &customersv1.GetCustomerRequest{Id: uuid.New().String()})
	if err == nil {
		t.Fatal(testWantErr)
	}
}

type stubGetInt struct{ grpcCustomerMock }

func (stubGetInt) Get(context.Context, string) (*domain.Customer, error) {
	return nil, context.Canceled
}
