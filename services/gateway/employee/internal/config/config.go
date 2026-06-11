package config

import (
	"github.com/dealer/dealer/pkg/configenv"
)

type Config struct {
	HTTPPort                    int
	AuthGRPCAddr                string
	CustomersGRPCAddr           string
	VehiclesGRPCAddr            string
	DealsGRPCAddr               string
	PartsGRPCAddr               string
	BrandsGRPCAddr              string
	DealerPointsGRPCAddr        string
	EmployeeStatisticsGRPCAddr  string
	ClientStatisticsGRPCAddr    string
	EmployeeReviewsGRPCAddr     string
	WorkOrdersGRPCAddr          string
	WorksGRPCAddr               string
	EmployeesGRPCAddr           string
}

func Load() *Config {
	return &Config{
		HTTPPort:             configenv.Int("GATEWAY_HTTP_PORT", 8090),
		AuthGRPCAddr:         configenv.String("AUTH_GRPC_ADDR", "127.0.0.1:50051"),
		CustomersGRPCAddr:    configenv.String("CUSTOMERS_GRPC_ADDR", "127.0.0.1:50052"),
		VehiclesGRPCAddr:     configenv.String("VEHICLES_GRPC_ADDR", "127.0.0.1:50053"),
		DealsGRPCAddr:        configenv.String("DEALS_GRPC_ADDR", "127.0.0.1:50054"),
		PartsGRPCAddr:        configenv.String("PARTS_GRPC_ADDR", "127.0.0.1:50055"),
		BrandsGRPCAddr:       configenv.String("BRANDS_GRPC_ADDR", "127.0.0.1:50056"),
		DealerPointsGRPCAddr:       configenv.String("DEALER_POINTS_GRPC_ADDR", "127.0.0.1:50057"),
		EmployeeStatisticsGRPCAddr: configenv.String("EMPLOYEE_STATISTICS_GRPC_ADDR", "127.0.0.1:50061"),
		ClientStatisticsGRPCAddr:   configenv.String("CLIENT_STATISTICS_GRPC_ADDR", "127.0.0.1:50062"),
		EmployeeReviewsGRPCAddr: configenv.String("EMPLOYEE_REVIEWS_GRPC_ADDR", "127.0.0.1:50063"),
		WorkOrdersGRPCAddr:      configenv.String("WORKORDERS_GRPC_ADDR", "127.0.0.1:50064"),
		WorksGRPCAddr:           configenv.String("WORKS_GRPC_ADDR", "127.0.0.1:50065"),
		EmployeesGRPCAddr:       configenv.String("EMPLOYEES_GRPC_ADDR", "127.0.0.1:50066"),
	}
}
