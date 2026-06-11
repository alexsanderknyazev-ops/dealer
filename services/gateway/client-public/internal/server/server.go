package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/dealer/dealer/pkg/grpclient"
	clientauthv1 "github.com/dealer/dealer/pkg/pb/clientauth/v1"
	clientsv1 "github.com/dealer/dealer/pkg/pb/clients/v1"
	"github.com/dealer/dealer/services/gateway/client-public/internal/config"
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
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
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
		{"client-auth-public", clientauthv1.RegisterClientAuthPublicServiceHandlerFromEndpoint, s.cfg.ClientAuthGRPCAddr},
		{"client-registration-public", clientsv1.RegisterClientRegistrationPublicServiceHandlerFromEndpoint, s.cfg.ClientRegistrationGRPCAddr},
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
