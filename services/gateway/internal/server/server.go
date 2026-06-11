package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/encoding/protojson"

	brandsv1 "github.com/dealer/dealer/pkg/pb/brands/v1"
	customersv1 "github.com/dealer/dealer/pkg/pb/customers/v1"
	dealerpointsv1 "github.com/dealer/dealer/pkg/pb/dealerpoints/v1"
	dealsv1 "github.com/dealer/dealer/pkg/pb/deals/v1"
	partsv1 "github.com/dealer/dealer/pkg/pb/parts/v1"
	vehiclesv1 "github.com/dealer/dealer/pkg/pb/vehicles/v1"
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
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	}
	registrars := []struct {
		name string
		fn   func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error
		addr string
	}{
		{"customers", customersv1.RegisterCustomersServiceHandlerFromEndpoint, s.cfg.CustomersGRPCAddr},
		{"vehicles", vehiclesv1.RegisterVehiclesServiceHandlerFromEndpoint, s.cfg.VehiclesGRPCAddr},
		{"deals", dealsv1.RegisterDealsServiceHandlerFromEndpoint, s.cfg.DealsGRPCAddr},
		{"parts", partsv1.RegisterPartsServiceHandlerFromEndpoint, s.cfg.PartsGRPCAddr},
		{"brands", brandsv1.RegisterBrandsServiceHandlerFromEndpoint, s.cfg.BrandsGRPCAddr},
		{"dealer-points", dealerpointsv1.RegisterDealerPointsServiceHandlerFromEndpoint, s.cfg.DealerPointsGRPCAddr},
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
