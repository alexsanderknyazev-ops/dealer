package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	brandsv1 "github.com/dealer/dealer/pkg/pb/brands/v1"
	"github.com/dealer/dealer/pkg/grpclient"
)

type BrandChecker struct {
	brands brandsv1.BrandsServiceClient
	conn   *grpc.ClientConn
}

func NewBrandChecker(ctx context.Context, brandsAddr string) (*BrandChecker, error) {
	conn, err := grpclient.Dial(ctx, brandsAddr)
	if err != nil {
		return nil, fmt.Errorf("dial brands %s: %w", brandsAddr, err)
	}
	return &BrandChecker{
		brands: brandsv1.NewBrandsServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *BrandChecker) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *BrandChecker) BrandExists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := c.brands.GetBrand(grpclient.OutgoingContext(ctx), &brandsv1.GetBrandRequest{Id: id.String()})
	return grpclient.ExistsFromRPC(err)
}
