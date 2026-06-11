package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	customersv1 "github.com/dealer/dealer/pkg/pb/customers/v1"
	vehiclesv1 "github.com/dealer/dealer/pkg/pb/vehicles/v1"
	"github.com/dealer/dealer/pkg/grpclient"
)

// ReferenceChecker validates foreign keys via customers/vehicles gRPC.
type ReferenceChecker struct {
	customers customersv1.CustomersServiceClient
	vehicles  vehiclesv1.VehiclesServiceClient
	conns     []*grpc.ClientConn
}

func NewReferenceChecker(ctx context.Context, customersAddr, vehiclesAddr string) (*ReferenceChecker, error) {
	cc, err := grpclient.Dial(ctx, customersAddr)
	if err != nil {
		return nil, fmt.Errorf("dial customers %s: %w", customersAddr, err)
	}
	vc, err := grpclient.Dial(ctx, vehiclesAddr)
	if err != nil {
		_ = cc.Close()
		return nil, fmt.Errorf("dial vehicles %s: %w", vehiclesAddr, err)
	}
	return &ReferenceChecker{
		customers: customersv1.NewCustomersServiceClient(cc),
		vehicles:  vehiclesv1.NewVehiclesServiceClient(vc),
		conns:     []*grpc.ClientConn{cc, vc},
	}, nil
}

func (c *ReferenceChecker) Close() error {
	var first error
	for _, conn := range c.conns {
		if err := conn.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (c *ReferenceChecker) CustomerExists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := c.customers.GetCustomer(grpclient.OutgoingContext(ctx), &customersv1.GetCustomerRequest{Id: id.String()})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *ReferenceChecker) VehicleExists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := c.vehicles.GetVehicle(grpclient.OutgoingContext(ctx), &vehiclesv1.GetVehicleRequest{Id: id.String()})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
