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
	worksv1 "github.com/dealer/dealer/pkg/pb/works/v1"
	"github.com/dealer/dealer/pkg/grpclient"
)

type ReferenceChecker struct {
	customers    customersv1.CustomersServiceClient
	vehicles     vehiclesv1.VehiclesServiceClient
	dealerPoints dealerpointsv1.DealerPointsServiceClient
	parts        partsv1.PartsServiceClient
	works        worksv1.WorksServiceClient
	conns        []*grpc.ClientConn
}

func NewReferenceChecker(ctx context.Context, customersAddr, vehiclesAddr, dealerPointsAddr, partsAddr, worksAddr string) (*ReferenceChecker, error) {
	rc := &ReferenceChecker{}
	dial := func(addr string) (*grpc.ClientConn, error) {
		if addr == "" {
			return nil, nil
		}
		return grpclient.Dial(ctx, addr)
	}
	for name, addr := range map[string]string{
		"customers": customersAddr, "vehicles": vehiclesAddr, "dealer-points": dealerPointsAddr,
		"parts": partsAddr, "works": worksAddr,
	} {
		cc, err := dial(addr)
		if err != nil {
			rc.Close()
			return nil, fmt.Errorf("dial %s: %w", name, err)
		}
		if cc == nil {
			continue
		}
		rc.conns = append(rc.conns, cc)
		switch name {
		case "customers":
			rc.customers = customersv1.NewCustomersServiceClient(cc)
		case "vehicles":
			rc.vehicles = vehiclesv1.NewVehiclesServiceClient(cc)
		case "dealer-points":
			rc.dealerPoints = dealerpointsv1.NewDealerPointsServiceClient(cc)
		case "parts":
			rc.parts = partsv1.NewPartsServiceClient(cc)
		case "works":
			rc.works = worksv1.NewWorksServiceClient(cc)
		}
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
	return v.Vin, fmt.Sprintf("%s %s %d", v.Make, v.Model, v.Year)
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

func (c *ReferenceChecker) WorkDisplay(ctx context.Context, id uuid.UUID) (code, name string) {
	if c.works == nil {
		return "", ""
	}
	resp, err := c.works.GetWork(grpclient.OutgoingContext(ctx), &worksv1.GetWorkRequest{Id: id.String()})
	if err != nil || resp.Work == nil {
		return "", ""
	}
	return resp.Work.Code, resp.Work.Name
}
