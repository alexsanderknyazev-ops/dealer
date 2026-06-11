package client

import (
	"context"

	vehiclesv1 "github.com/dealer/dealer/pkg/pb/vehicles/v1"
	"github.com/dealer/dealer/pkg/grpclient"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type VehiclesClient struct {
	conn *grpc.ClientConn
	api  vehiclesv1.VehiclesServiceClient
}

func NewVehiclesClient(ctx context.Context, addr string) (*VehiclesClient, error) {
	conn, err := grpclient.Dial(ctx, addr)
	if err != nil {
		return nil, err
	}
	return &VehiclesClient{conn: conn, api: vehiclesv1.NewVehiclesServiceClient(conn)}, nil
}

func (c *VehiclesClient) Close() error {
	return c.conn.Close()
}

func (c *VehiclesClient) GetByID(ctx context.Context, id string) (*vehiclesv1.Vehicle, error) {
	resp, err := c.api.GetVehicle(grpclient.OutgoingContext(ctx), &vehiclesv1.GetVehicleRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return resp.Vehicle, nil
}

func IsNotFound(err error) bool {
	return status.Code(err) == codes.NotFound
}
