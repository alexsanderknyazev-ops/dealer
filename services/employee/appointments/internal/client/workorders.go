package client

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/grpc"

	workordersv1 "github.com/dealer/dealer/pkg/pb/workorders/v1"
	"github.com/dealer/dealer/pkg/grpclient"
	"github.com/dealer/dealer/services/appointments/internal/domain"
)

type WorkOrdersCreator struct {
	client workordersv1.WorkOrdersServiceClient
	conn   *grpc.ClientConn
}

func NewWorkOrdersCreator(ctx context.Context, addr string) (*WorkOrdersCreator, error) {
	if addr == "" {
		return nil, nil
	}
	cc, err := grpclient.Dial(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("dial workorders: %w", err)
	}
	return &WorkOrdersCreator{
		client: workordersv1.NewWorkOrdersServiceClient(cc),
		conn:   cc,
	}, nil
}

func (c *WorkOrdersCreator) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *WorkOrdersCreator) CreateFromAppointment(ctx context.Context, a *domain.RepairAppointment) (id, number string, err error) {
	if c == nil {
		return "", "", fmt.Errorf("workorders service unavailable")
	}
	labor := make([]*workordersv1.WorkOrderLaborInput, len(a.Labor))
	for i, l := range a.Labor {
		item := &workordersv1.WorkOrderLaborInput{
			Description: l.Description,
			Quantity:    l.Quantity,
			UnitPrice:   l.UnitPrice,
			SortOrder:   l.SortOrder,
		}
		if l.WorkID != nil {
			item.WorkId = l.WorkID.String()
		}
		labor[i] = item
	}
	parts := make([]*workordersv1.WorkOrderPartInput, len(a.Parts))
	for i, p := range a.Parts {
		parts[i] = &workordersv1.WorkOrderPartInput{
			PartId:      p.PartID.String(),
			WarehouseId: p.WarehouseID.String(),
			Quantity:    strconv.Itoa(int(p.Quantity)),
			UnitPrice:   p.UnitPrice,
			Description: p.Notes,
			SortOrder:   p.SortOrder,
		}
	}
	req := &workordersv1.CreateWorkOrderRequest{
		CustomerId:  a.CustomerID.String(),
		VehicleId:   a.VehicleID.String(),
		RepairType:  "commercial",
		Status:      "draft",
		Complaint:   a.Complaint,
		Notes:       fmt.Sprintf("Из записи на ремонт %s", a.AppointmentNumber),
		Labor:       labor,
		Parts:       parts,
	}
	if a.DealerPointID != nil {
		req.DealerPointId = a.DealerPointID.String()
	}
	if a.WarehouseID != nil {
		req.WarehouseId = a.WarehouseID.String()
	}
	if a.ScheduledStart.Unix() > 0 {
		req.OpenedAt = a.ScheduledStart.Unix()
	}
	resp, err := c.client.CreateWorkOrder(grpclient.OutgoingContext(ctx), req)
	if err != nil {
		return "", "", err
	}
	if resp.WorkOrder == nil {
		return "", "", fmt.Errorf("empty work order response")
	}
	return resp.WorkOrder.Id, resp.WorkOrder.OrderNumber, nil
}

func (c *WorkOrdersCreator) GetOrderNumber(ctx context.Context, id string) string {
	if c == nil {
		return ""
	}
	resp, err := c.client.GetWorkOrder(grpclient.OutgoingContext(ctx), &workordersv1.GetWorkOrderRequest{Id: id})
	if err != nil || resp.WorkOrder == nil {
		return ""
	}
	return resp.WorkOrder.OrderNumber
}
