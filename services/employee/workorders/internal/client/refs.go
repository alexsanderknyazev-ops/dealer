package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	customersv1 "github.com/dealer/dealer/pkg/pb/customers/v1"
	dealerpointsv1 "github.com/dealer/dealer/pkg/pb/dealerpoints/v1"
	partsv1 "github.com/dealer/dealer/pkg/pb/parts/v1"
	vehiclesv1 "github.com/dealer/dealer/pkg/pb/vehicles/v1"
	employeesv1 "github.com/dealer/dealer/pkg/pb/employees/v1"
	worksv1 "github.com/dealer/dealer/pkg/pb/works/v1"
	"github.com/dealer/dealer/pkg/grpclient"
	"github.com/dealer/dealer/services/workorders/internal/domain"
)

type ReferenceChecker struct {
	customers    customersv1.CustomersServiceClient
	vehicles     vehiclesv1.VehiclesServiceClient
	dealerPoints dealerpointsv1.DealerPointsServiceClient
	parts        partsv1.PartsServiceClient
	works        worksv1.WorksServiceClient
	employees    employeesv1.EmployeesServiceClient
	conns        []*grpc.ClientConn
}

func NewReferenceChecker(
	ctx context.Context,
	customersAddr, vehiclesAddr, dealerPointsAddr, partsAddr, worksAddr, employeesAddr string,
) (*ReferenceChecker, error) {
	rc := &ReferenceChecker{}
	dial := func(addr string) (*grpc.ClientConn, error) {
		if addr == "" {
			return nil, nil
		}
		return grpclient.Dial(ctx, addr)
	}
	cc, err := dial(customersAddr)
	if err != nil {
		return nil, fmt.Errorf("dial customers: %w", err)
	}
	if cc != nil {
		rc.conns = append(rc.conns, cc)
		rc.customers = customersv1.NewCustomersServiceClient(cc)
	}
	vc, err := dial(vehiclesAddr)
	if err != nil {
		rc.Close()
		return nil, fmt.Errorf("dial vehicles: %w", err)
	}
	if vc != nil {
		rc.conns = append(rc.conns, vc)
		rc.vehicles = vehiclesv1.NewVehiclesServiceClient(vc)
	}
	dpc, err := dial(dealerPointsAddr)
	if err != nil {
		rc.Close()
		return nil, fmt.Errorf("dial dealer-points: %w", err)
	}
	if dpc != nil {
		rc.conns = append(rc.conns, dpc)
		rc.dealerPoints = dealerpointsv1.NewDealerPointsServiceClient(dpc)
	}
	pc, err := dial(partsAddr)
	if err != nil {
		rc.Close()
		return nil, fmt.Errorf("dial parts: %w", err)
	}
	if pc != nil {
		rc.conns = append(rc.conns, pc)
		rc.parts = partsv1.NewPartsServiceClient(pc)
	}
	wc, err := dial(worksAddr)
	if err != nil {
		rc.Close()
		return nil, fmt.Errorf("dial works: %w", err)
	}
	if wc != nil {
		rc.conns = append(rc.conns, wc)
		rc.works = worksv1.NewWorksServiceClient(wc)
	}
	ec, err := dial(employeesAddr)
	if err != nil {
		rc.Close()
		return nil, fmt.Errorf("dial employees: %w", err)
	}
	if ec != nil {
		rc.conns = append(rc.conns, ec)
		rc.employees = employeesv1.NewEmployeesServiceClient(ec)
	}
	return rc, nil
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
	if c.customers == nil {
		return true, nil
	}
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
	if c.vehicles == nil {
		return true, nil
	}
	_, err := c.vehicles.GetVehicle(grpclient.OutgoingContext(ctx), &vehiclesv1.GetVehicleRequest{Id: id.String()})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *ReferenceChecker) DealerPointExists(ctx context.Context, id uuid.UUID) (bool, error) {
	if c.dealerPoints == nil {
		return true, nil
	}
	_, err := c.dealerPoints.GetDealerPoint(grpclient.OutgoingContext(ctx), &dealerpointsv1.GetDealerPointRequest{Id: id.String()})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *ReferenceChecker) WarehouseExists(ctx context.Context, id uuid.UUID) (bool, error) {
	if c.dealerPoints == nil {
		return true, nil
	}
	_, err := c.dealerPoints.GetWarehouse(grpclient.OutgoingContext(ctx), &dealerpointsv1.GetWarehouseRequest{Id: id.String()})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *ReferenceChecker) WorkExists(ctx context.Context, id uuid.UUID) (bool, error) {
	if c.works == nil {
		return true, nil
	}
	_, err := c.works.GetWork(grpclient.OutgoingContext(ctx), &worksv1.GetWorkRequest{Id: id.String()})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *ReferenceChecker) EmployeeExists(ctx context.Context, id uuid.UUID) (bool, error) {
	if c.employees == nil {
		return true, nil
	}
	_, err := c.employees.GetEmployee(grpclient.OutgoingContext(ctx), &employeesv1.GetEmployeeRequest{Id: id.String()})
	if err == nil {
		return true, nil
	}
	if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
		_, err = c.employees.GetEmployeeByUserID(grpclient.OutgoingContext(ctx), &employeesv1.GetEmployeeByUserIDRequest{UserId: id.String()})
		if err == nil {
			return true, nil
		}
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	return false, err
}

func (c *ReferenceChecker) EmployeeFullName(ctx context.Context, id uuid.UUID) string {
	if c.employees == nil {
		return ""
	}
	resp, err := c.employees.GetEmployee(grpclient.OutgoingContext(ctx), &employeesv1.GetEmployeeRequest{Id: id.String()})
	if err == nil && resp.Employee != nil {
		return resp.Employee.FullName
	}
	resp, err = c.employees.GetEmployeeByUserID(grpclient.OutgoingContext(ctx), &employeesv1.GetEmployeeByUserIDRequest{UserId: id.String()})
	if err != nil || resp.Employee == nil {
		return ""
	}
	return resp.Employee.FullName
}

func (c *ReferenceChecker) ResolveWork(ctx context.Context, id uuid.UUID) (string, string, string, error) {
	if c.works == nil {
		return "", "", "", fmt.Errorf("works service unavailable")
	}
	resp, err := c.works.GetWork(grpclient.OutgoingContext(ctx), &worksv1.GetWorkRequest{Id: id.String()})
	if err != nil {
		return "", "", "", err
	}
	if resp.Work == nil {
		return "", "", "", fmt.Errorf("work not found")
	}
	return resp.Work.Name, resp.Work.LaborHours, resp.Work.UnitPrice, nil
}

func (c *ReferenceChecker) PartExists(ctx context.Context, id uuid.UUID) (bool, error) {
	if c.parts == nil {
		return true, nil
	}
	_, err := c.parts.GetPart(grpclient.OutgoingContext(ctx), &partsv1.GetPartRequest{Id: id.String()})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *ReferenceChecker) CustomerName(ctx context.Context, id uuid.UUID) string {
	if c.customers == nil {
		return ""
	}
	resp, err := c.customers.GetCustomer(grpclient.OutgoingContext(ctx), &customersv1.GetCustomerRequest{Id: id.String()})
	if err != nil || resp.Customer == nil {
		return ""
	}
	return resp.Customer.Name
}

func (c *ReferenceChecker) VehicleDisplay(ctx context.Context, id uuid.UUID) (vin, label string) {
	if c.vehicles == nil {
		return "", ""
	}
	resp, err := c.vehicles.GetVehicle(grpclient.OutgoingContext(ctx), &vehiclesv1.GetVehicleRequest{Id: id.String()})
	if err != nil || resp.Vehicle == nil {
		return "", ""
	}
	v := resp.Vehicle
	vin = v.Vin
	label = fmt.Sprintf("%s %s %d", v.Make, v.Model, v.Year)
	return vin, label
}

func (c *ReferenceChecker) PartDisplay(ctx context.Context, id uuid.UUID) (name, sku string) {
	if c.parts == nil {
		return "", ""
	}
	resp, err := c.parts.GetPart(grpclient.OutgoingContext(ctx), &partsv1.GetPartRequest{Id: id.String()})
	if err != nil || resp.Part == nil {
		return "", ""
	}
	return resp.Part.Name, resp.Part.Sku
}

func (c *ReferenceChecker) WarehouseName(ctx context.Context, id uuid.UUID) string {
	if c.dealerPoints == nil {
		return ""
	}
	resp, err := c.dealerPoints.GetWarehouse(grpclient.OutgoingContext(ctx), &dealerpointsv1.GetWarehouseRequest{Id: id.String()})
	if err != nil || resp.Warehouse == nil {
		return ""
	}
	return resp.Warehouse.Name
}

func (c *ReferenceChecker) CreateMovementDocument(
	ctx context.Context,
	workOrderID uuid.UUID,
	orderNumber string,
	lines []domain.MovementDocumentLineInput,
	createdBy string,
) (string, error) {
	if c.parts == nil {
		return "", fmt.Errorf("parts service unavailable")
	}
	protoLines := make([]*partsv1.MovementDocumentLineInput, len(lines))
	for i, it := range lines {
		protoLines[i] = &partsv1.MovementDocumentLineInput{
			PartId:          it.PartID.String(),
			WarehouseId:     it.WarehouseID.String(),
			Quantity:        it.Quantity,
			ReferenceLineId: it.LineID.String(),
			Notes:           it.Notes,
			SortOrder:       it.SortOrder,
		}
	}
	resp, err := c.parts.CreateMovementDocument(grpclient.OutgoingContext(ctx), &partsv1.CreateMovementDocumentRequest{
		MovementType:  "work_order_issue",
		ReferenceType: "work_order",
		ReferenceId:   workOrderID.String(),
		Notes:         fmt.Sprintf("Заказ-наряд %s", orderNumber),
		CreatedBy:     createdBy,
		Lines:         protoLines,
	})
	if err != nil {
		return "", err
	}
	if resp.Document == nil {
		return "", fmt.Errorf("empty movement document response")
	}
	return resp.Document.Id, nil
}
