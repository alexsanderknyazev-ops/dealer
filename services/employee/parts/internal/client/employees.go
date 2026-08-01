package client

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	employeesv1 "github.com/dealer/dealer/pkg/pb/employees/v1"
	"github.com/dealer/dealer/pkg/grpclient"
)

type EmployeeResolver struct {
	client employeesv1.EmployeesServiceClient
	conn   *grpc.ClientConn
}

func NewEmployeeResolver(ctx context.Context, addr string) (*EmployeeResolver, error) {
	if addr == "" {
		return nil, nil
	}
	conn, err := grpclient.Dial(ctx, addr)
	if err != nil {
		return nil, err
	}
	return &EmployeeResolver{
		client: employeesv1.NewEmployeesServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *EmployeeResolver) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *EmployeeResolver) FullName(ctx context.Context, ref uuid.UUID) string {
	if c == nil || c.client == nil {
		return ""
	}
	resp, err := c.client.GetEmployee(grpclient.OutgoingContext(ctx), &employeesv1.GetEmployeeRequest{Id: ref.String()})
	if err == nil && resp.Employee != nil {
		return resp.Employee.FullName
	}
	resp, err = c.client.GetEmployeeByUserID(grpclient.OutgoingContext(ctx), &employeesv1.GetEmployeeByUserIDRequest{UserId: ref.String()})
	if err != nil {
		return ""
	}
	if resp.Employee == nil {
		return ""
	}
	return resp.Employee.FullName
}

func (c *EmployeeResolver) Exists(ctx context.Context, ref uuid.UUID) (bool, error) {
	if c == nil || c.client == nil {
		return true, nil
	}
	_, err := c.client.GetEmployee(grpclient.OutgoingContext(ctx), &employeesv1.GetEmployeeRequest{Id: ref.String()})
	if err == nil {
		return true, nil
	}
	if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
		_, err = c.client.GetEmployeeByUserID(grpclient.OutgoingContext(ctx), &employeesv1.GetEmployeeByUserIDRequest{UserId: ref.String()})
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
