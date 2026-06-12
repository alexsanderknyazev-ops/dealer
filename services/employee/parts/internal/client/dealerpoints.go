package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	dealerpointsv1 "github.com/dealer/dealer/pkg/pb/dealerpoints/v1"
	"github.com/dealer/dealer/pkg/grpclient"
)

type DealerPointsChecker struct {
	dealerPoints dealerpointsv1.DealerPointsServiceClient
	conn         *grpc.ClientConn
}

func NewDealerPointsChecker(ctx context.Context, addr string) (*DealerPointsChecker, error) {
	conn, err := grpclient.Dial(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("dial dealer-points %s: %w", addr, err)
	}
	return &DealerPointsChecker{
		dealerPoints: dealerpointsv1.NewDealerPointsServiceClient(conn),
		conn:         conn,
	}, nil
}

func (c *DealerPointsChecker) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *DealerPointsChecker) DealerPointExists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := c.dealerPoints.GetDealerPoint(grpclient.OutgoingContext(ctx), &dealerpointsv1.GetDealerPointRequest{Id: id.String()})
	return grpclient.ExistsFromRPC(err)
}

func (c *DealerPointsChecker) LegalEntityExists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := c.dealerPoints.GetLegalEntity(grpclient.OutgoingContext(ctx), &dealerpointsv1.GetLegalEntityRequest{Id: id.String()})
	return grpclient.ExistsFromRPC(err)
}

func (c *DealerPointsChecker) WarehouseExists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := c.dealerPoints.GetWarehouse(grpclient.OutgoingContext(ctx), &dealerpointsv1.GetWarehouseRequest{Id: id.String()})
	return grpclient.ExistsFromRPC(err)
}

func (c *DealerPointsChecker) WarehouseName(ctx context.Context, id uuid.UUID) string {
	resp, err := c.dealerPoints.GetWarehouse(grpclient.OutgoingContext(ctx), &dealerpointsv1.GetWarehouseRequest{Id: id.String()})
	if err != nil || resp.Warehouse == nil {
		return ""
	}
	return resp.Warehouse.Name
}
