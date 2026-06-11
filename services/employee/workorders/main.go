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
	workordersv1 "github.com/dealer/dealer/pkg/pb/workorders/v1"
	"github.com/dealer/dealer/pkg/postgres"
	"github.com/dealer/dealer/services/workorders/internal/client"
	"github.com/dealer/dealer/services/workorders/internal/config"
	grpcserver "github.com/dealer/dealer/services/workorders/internal/grpc"
	"github.com/dealer/dealer/services/workorders/internal/repository"
	"github.com/dealer/dealer/services/workorders/internal/service"
)

const serviceName = "workorders-service"

func main() {
	cfg := config.Load()
	logger := observe.Init(serviceName)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN, dbschema.WorkOrders, dbschema.Public)
	if err != nil {
		logger.Error("postgres connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.NewWorkOrderRepository(pool)
	var refs service.ReferenceChecker
	if cfg.CustomersGRPCAddr != "" && cfg.VehiclesGRPCAddr != "" {
		var refClient *client.ReferenceChecker
		for attempt := 1; attempt <= 6; attempt++ {
			dialCtx, dialCancel := context.WithTimeout(ctx, 10*time.Second)
			refClient, err = client.NewReferenceChecker(
				dialCtx,
				cfg.CustomersGRPCAddr,
				cfg.VehiclesGRPCAddr,
				cfg.DealerPointsGRPCAddr,
				cfg.PartsGRPCAddr,
				cfg.WorksGRPCAddr,
				cfg.EmployeesGRPCAddr,
			)
			dialCancel()
			if err == nil {
				break
			}
			if attempt < 6 {
				logger.Warn("gRPC ref clients connect failed; retrying", "attempt", attempt, "err", err)
				time.Sleep(3 * time.Second)
			}
		}
		if err != nil {
			logger.Warn("gRPC ref clients connect failed; reference validation and movement docs disabled until restart", "err", err)
		} else {
			defer refClient.Close()
			refs = refClient
			logger.Info("reference checks via gRPC enabled")
		}
	} else {
		logger.Warn("reference gRPC addrs not set; validation disabled")
	}
	svc := service.NewWorkOrderService(repo, refs)

	var employeeNamer grpcserver.EmployeeNamer
	if rc, ok := refs.(*client.ReferenceChecker); ok {
		employeeNamer = rc
	}

	gsrv := grpc.NewServer(observe.GRPCServerOptions(serviceName, logger, &grpcauth.Config{
		JWTSecret:  cfg.JWTSecret,
		WriteRoles: []string{"admin", "manager", "master", "service_advisor", "storekeeper", "parts_manager"},
	})...)
	workordersv1.RegisterWorkOrdersServiceServer(gsrv, grpcserver.NewServer(svc, employeeNamer))
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
