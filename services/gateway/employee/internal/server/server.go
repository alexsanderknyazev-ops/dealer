package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/dealer/dealer/pkg/grpclient"
	authv1 "github.com/dealer/dealer/pkg/pb/auth/v1"
	brandsv1 "github.com/dealer/dealer/pkg/pb/brands/v1"
	customersv1 "github.com/dealer/dealer/pkg/pb/customers/v1"
	dealerpointsv1 "github.com/dealer/dealer/pkg/pb/dealerpoints/v1"
	dealsv1 "github.com/dealer/dealer/pkg/pb/deals/v1"
	partsv1 "github.com/dealer/dealer/pkg/pb/parts/v1"
	clientstatsv1 "github.com/dealer/dealer/pkg/pb/statistics/client/v1"
	employeestatsv1 "github.com/dealer/dealer/pkg/pb/statistics/employee/v1"
	reviewsv1 "github.com/dealer/dealer/pkg/pb/reviews/v1"
	vehiclesv1 "github.com/dealer/dealer/pkg/pb/vehicles/v1"
	workordersv1 "github.com/dealer/dealer/pkg/pb/workorders/v1"
	employeesv1 "github.com/dealer/dealer/pkg/pb/employees/v1"
	worksv1 "github.com/dealer/dealer/pkg/pb/works/v1"
	"github.com/dealer/dealer/services/gateway/internal/config"
)

type Server struct {
	cfg *config.Config
	mux *runtime.ServeMux
}

func New(cfg *config.Config) (*Server, error) {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true, // keep list fields (deals: [], total: 0) for frontend contract
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			switch key {
			case "Authorization":
				return "authorization", true
			default:
				return runtime.DefaultHeaderMatcher(key)
			}
		}),
	)

	s := &Server{cfg: cfg, mux: mux}
	if err := s.registerBackends(context.Background()); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Server) Handler() http.Handler {
	return cors(s.mux)
}

func (s *Server) registerBackends(ctx context.Context) error {
	opts := grpclient.DefaultDialOptions()
	registrars := []struct {
		name string
		fn   func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error
		addr string
	}{
		{"auth", authv1.RegisterAuthServiceHandlerFromEndpoint, s.cfg.AuthGRPCAddr},
		{"customers", customersv1.RegisterCustomersServiceHandlerFromEndpoint, s.cfg.CustomersGRPCAddr},
		{"vehicles", vehiclesv1.RegisterVehiclesServiceHandlerFromEndpoint, s.cfg.VehiclesGRPCAddr},
		{"deals", dealsv1.RegisterDealsServiceHandlerFromEndpoint, s.cfg.DealsGRPCAddr},
		{"parts", partsv1.RegisterPartsServiceHandlerFromEndpoint, s.cfg.PartsGRPCAddr},
		{"brands", brandsv1.RegisterBrandsServiceHandlerFromEndpoint, s.cfg.BrandsGRPCAddr},
		{"dealer-points", dealerpointsv1.RegisterDealerPointsServiceHandlerFromEndpoint, s.cfg.DealerPointsGRPCAddr},
		{"employee-statistics", employeestatsv1.RegisterEmployeeStatisticsServiceHandlerFromEndpoint, s.cfg.EmployeeStatisticsGRPCAddr},
		{"client-statistics", clientstatsv1.RegisterClientStatisticsServiceHandlerFromEndpoint, s.cfg.ClientStatisticsGRPCAddr},
		{"employee-reviews", reviewsv1.RegisterEmployeeReviewsServiceHandlerFromEndpoint, s.cfg.EmployeeReviewsGRPCAddr},
		{"work-orders", workordersv1.RegisterWorkOrdersServiceHandlerFromEndpoint, s.cfg.WorkOrdersGRPCAddr},
		{"works", worksv1.RegisterWorksServiceHandlerFromEndpoint, s.cfg.WorksGRPCAddr},
		{"employees", employeesv1.RegisterEmployeesServiceHandlerFromEndpoint, s.cfg.EmployeesGRPCAddr},
	}
	for _, r := range registrars {
		if err := r.fn(ctx, s.mux, r.addr, opts); err != nil {
			return fmt.Errorf("register %s gateway: %w", r.name, err)
		}
	}
	return nil
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
