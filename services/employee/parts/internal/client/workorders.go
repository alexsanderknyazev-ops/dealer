package client

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	workordersv1 "github.com/dealer/dealer/pkg/pb/workorders/v1"
	"github.com/dealer/dealer/pkg/grpclient"
)

type WorkOrdersNotifier struct {
	client workordersv1.WorkOrdersServiceClient
	conn   *grpc.ClientConn
}

func NewWorkOrdersNotifier(ctx context.Context, addr string) (*WorkOrdersNotifier, error) {
	if addr == "" {
		return nil, nil
	}
	cc, err := grpclient.Dial(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("dial workorders %s: %w", addr, err)
	}
	return &WorkOrdersNotifier{
		client: workordersv1.NewWorkOrdersServiceClient(cc),
		conn:   cc,
	}, nil
}

func (c *WorkOrdersNotifier) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *WorkOrdersNotifier) ApplyMovementDocument(ctx context.Context, workOrderID, documentID, status string) error {
	if c == nil {
		return nil
	}
	_, err := c.client.ApplyMovementDocument(grpclient.OutgoingContext(ctx), &workordersv1.ApplyMovementDocumentRequest{
		WorkOrderId:            workOrderID,
		MovementDocumentId:     documentID,
		MovementDocumentStatus: status,
	})
	return err
}

func (c *WorkOrdersNotifier) GetWorkOrder(ctx context.Context, id string) (*workordersv1.WorkOrder, error) {
	if c == nil {
		return nil, fmt.Errorf("workorders client unavailable")
	}
	resp, err := c.client.GetWorkOrder(grpclient.OutgoingContext(ctx), &workordersv1.GetWorkOrderRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return resp.WorkOrder, nil
}
