package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/dealer/dealer/pkg/dbschema"
	"github.com/dealer/dealer/pkg/grpcauth"
	"github.com/dealer/dealer/pkg/health"
	"github.com/dealer/dealer/pkg/observe"
	vehiclesv1 "github.com/dealer/dealer/pkg/pb/vehicles/v1"
	"github.com/dealer/dealer/pkg/postgres"
	"github.com/dealer/dealer/services/vehicles/internal/client"
	"github.com/dealer/dealer/services/vehicles/internal/config"
	grpcserver "github.com/dealer/dealer/services/vehicles/internal/grpc"
	"github.com/dealer/dealer/services/vehicles/internal/repository"
	"github.com/dealer/dealer/services/vehicles/internal/service"
)

const serviceName = "vehicles-service"

func main() {
	cfg := config.Load()
	logger := observe.Init(serviceName)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN, dbschema.Vehicles, dbschema.Public)
	if err != nil {
		logger.Error("postgres connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.NewVehicleRepository(pool)
	var brands service.BrandChecker
	if cfg.BrandsGRPCAddr != "" {
		dialCtx, dialCancel := context.WithTimeout(ctx, 10*time.Second)
		brandClient, err := client.NewBrandChecker(dialCtx, cfg.BrandsGRPCAddr)
		dialCancel()
		if err != nil {
			logger.Error("gRPC brands client connect failed", "err", err)
			os.Exit(1)
		}
		defer brandClient.Close()
		brands = brandClient
		logger.Info("brand checks via gRPC", "brands", cfg.BrandsGRPCAddr)
	} else {
		logger.Warn("BRANDS_GRPC_ADDR not set; vehicle brand checks disabled")
	}
	var dealerPoints service.DealerPointsChecker
	if cfg.DealerPointsGRPCAddr != "" {
		dialCtx, dialCancel := context.WithTimeout(ctx, 10*time.Second)
		dpClient, err := client.NewDealerPointsChecker(dialCtx, cfg.DealerPointsGRPCAddr)
		dialCancel()
		if err != nil {
			logger.Error("gRPC dealer-points client connect failed", "err", err)
			os.Exit(1)
		}
		defer dpClient.Close()
		dealerPoints = dpClient
		logger.Info("dealer point checks via gRPC", "dealer_points", cfg.DealerPointsGRPCAddr)
	} else {
		logger.Warn("DEALER_POINTS_GRPC_ADDR not set; vehicle dealer point checks disabled")
	}
	svc := service.NewVehicleService(repo, brands, dealerPoints)

	gsrv := grpc.NewServer(observe.GRPCServerOptions(serviceName, logger, &grpcauth.Config{
		JWTSecret:  cfg.JWTSecret,
		WriteRoles: []string{"admin", "manager", "sales"},
		PublicMethods: []string{
			"/vehicles.v1.VehiclesService/GetVehicleByVIN",
			"/vehicles.v1.VehiclesService/GetVehicle",
		},
	})...)
	vehiclesv1.RegisterVehiclesServiceServer(gsrv, grpcserver.NewServer(svc))
	reflection.Register(gsrv)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		logger.Error("grpc listen failed", "err", err)
		os.Exit(1)
	}
	go func() {
		logger.Info("gRPC listening", "port", cfg.GRPCPort)
		if err := gsrv.Serve(lis); err != nil {
			logger.Error("grpc serve failed", "err", err)
		}
	}()

	httpMux := http.NewServeMux()
	observe.RegisterHTTP(httpMux, health.Postgres(pool))
	httpHandler := observe.WrapHTTP(serviceName, httpMux, logger)
	httpLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.HTTPPort))
	if err != nil {
		logger.Error("http listen failed", "err", err)
		os.Exit(1)
	}
	go func() {
		logger.Info("HTTP listening", "port", cfg.HTTPPort)
		if err := http.Serve(httpLis, httpHandler); err != nil {
			logger.Error("http serve failed", "err", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	stop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	done := make(chan struct{})
	go func() {
		gsrv.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
	case <-shutdownCtx.Done():
		gsrv.Stop()
	}
}
