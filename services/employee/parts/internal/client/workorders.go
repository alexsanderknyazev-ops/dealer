package client

import (
	"context"
	"fmt"
	"google.golang.org/grpc"

	workordersv1 "github.com/dealer/dealer/pkg/pb/workorders/v1"
	"github.com/dealer/dealer/pkg/grpclient"
	"github.com/dealer/dealer/services/parts/internal/service"
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

func (c *WorkOrdersNotifier) CreateWorkOrder(ctx context.Context, in service.CreateWorkOrderFromOrderInput) (*service.CreatedWorkOrderRef, error) {
	if c == nil {
		return nil, fmt.Errorf("workorders client unavailable")
	}
	parts := make([]*workordersv1.WorkOrderPartInput, len(in.Parts))
	for i, p := range in.Parts {
		parts[i] = &workordersv1.WorkOrderPartInput{
			PartId:      p.PartID,
			WarehouseId: p.WarehouseID,
			Quantity:    p.Quantity,
			UnitPrice:   p.UnitPrice,
			SortOrder:   p.SortOrder,
		}
	}
	resp, err := c.client.CreateWorkOrder(grpclient.OutgoingContext(ctx), &workordersv1.CreateWorkOrderRequest{
		CustomerId:       in.CustomerID,
		VehicleId:        in.VehicleID,
		WarehouseId:      in.WarehouseID,
		RepairType:       "commercial",
		Status:           "draft",
		Notes:            in.Notes,
		Parts:            parts,
		SourceOrderType:  in.SourceOrderType,
		SourceOrderId:    in.SourceOrderID,
	})
	if err != nil {
		return nil, err
	}
	if resp.WorkOrder == nil {
		return nil, fmt.Errorf("empty work order response")
	}
	return &service.CreatedWorkOrderRef{ID: resp.WorkOrder.Id, OrderNumber: resp.WorkOrder.OrderNumber}, nil
}

func (c *WorkOrdersNotifier) GetWorkOrder(ctx context.Context, id string) (string, error) {
	wo, err := c.GetWorkOrderDetails(ctx, id)
	if err != nil {
		return "", err
	}
	return wo.OrderNumber, nil
}

func (c *WorkOrdersNotifier) GetWorkOrderDetails(ctx context.Context, id string) (*workordersv1.WorkOrder, error) {
	if c == nil {
		return nil, fmt.Errorf("workorders client unavailable")
	}
	resp, err := c.client.GetWorkOrder(grpclient.OutgoingContext(ctx), &workordersv1.GetWorkOrderRequest{Id: id})
	if err != nil {
		return nil, err
	}
	if resp.WorkOrder == nil {
		return nil, fmt.Errorf("work order not found")
	}
	return resp.WorkOrder, nil
}
