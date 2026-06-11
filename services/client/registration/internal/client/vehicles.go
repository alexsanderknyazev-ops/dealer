package client

import (
	"context"

	vehiclesv1 "github.com/dealer/dealer/pkg/pb/vehicles/v1"
	"github.com/dealer/dealer/pkg/grpclient"
	"google.golang.org/grpc"
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

func (c *VehiclesClient) GetByVIN(ctx context.Context, vin string) (*vehiclesv1.Vehicle, error) {
	resp, err := c.api.GetVehicleByVIN(ctx, &vehiclesv1.GetVehicleByVINRequest{Vin: vin})
	if err != nil {
		return nil, err
	}
	return resp.Vehicle, nil
}
