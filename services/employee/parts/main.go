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
	partsv1 "github.com/dealer/dealer/pkg/pb/parts/v1"
	"github.com/dealer/dealer/pkg/postgres"
	"github.com/dealer/dealer/services/parts/internal/client"
	"github.com/dealer/dealer/services/parts/internal/config"
	grpcserver "github.com/dealer/dealer/services/parts/internal/grpc"
	"github.com/dealer/dealer/services/parts/internal/repository"
	"github.com/dealer/dealer/services/parts/internal/service"
)

const serviceName = "parts-service"

func main() {
	cfg := config.Load()
	logger := observe.Init(serviceName)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN, dbschema.Parts, dbschema.Public)
	if err != nil {
		logger.Error("postgres connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.NewPartRepository(pool)
	folderRepo := repository.NewFolderRepository(pool)
	stockRepo := repository.NewPartStockRepository(pool)
	movementRepo := repository.NewStockMovementRepository(pool)
	movementDocRepo := repository.NewMovementDocumentRepository(pool)
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
		logger.Warn("BRANDS_GRPC_ADDR not set; parts brand checks disabled")
	}
	var dealerPoints service.DealerPointsChecker
	var dpChecker *client.DealerPointsChecker
	if cfg.DealerPointsGRPCAddr != "" {
		dialCtx, dialCancel := context.WithTimeout(ctx, 10*time.Second)
		dpClient, err := client.NewDealerPointsChecker(dialCtx, cfg.DealerPointsGRPCAddr)
		dialCancel()
		if err != nil {
			logger.Error("gRPC dealer-points client connect failed", "err", err)
			os.Exit(1)
		}
		defer dpClient.Close()
		dpChecker = dpClient
		dealerPoints = dpClient
		logger.Info("dealer point checks via gRPC", "dealer_points", cfg.DealerPointsGRPCAddr)
	} else {
		logger.Warn("DEALER_POINTS_GRPC_ADDR not set; parts dealer point checks disabled")
	}
	var workOrders service.WorkOrdersNotifier
	var woNotifier *client.WorkOrdersNotifier
	if cfg.WorkOrdersGRPCAddr != "" {
		// Short timeout: workorders may start after parts; blocking here delays gRPC listen and breaks workorders startup.
		dialCtx, dialCancel := context.WithTimeout(ctx, 2*time.Second)
		woClient, err := client.NewWorkOrdersNotifier(dialCtx, cfg.WorkOrdersGRPCAddr)
		dialCancel()
		if err != nil {
			logger.Warn("gRPC workorders client connect failed; confirm won't update work orders until restart", "err", err)
		} else {
			defer woClient.Close()
			woNotifier = woClient
			workOrders = woClient
			logger.Info("work order notify via gRPC", "workorders", cfg.WorkOrdersGRPCAddr)
		}
	} else {
		logger.Warn("WORKORDERS_GRPC_ADDR not set; movement confirm won't update work orders")
	}
	var employees *client.EmployeeResolver
	if cfg.EmployeesGRPCAddr != "" {
		dialCtx, dialCancel := context.WithTimeout(ctx, 2*time.Second)
		empClient, err := client.NewEmployeeResolver(dialCtx, cfg.EmployeesGRPCAddr)
		dialCancel()
		if err != nil {
			logger.Warn("gRPC employees client connect failed; employee names disabled until restart", "err", err)
		} else if empClient != nil {
			defer empClient.Close()
			employees = empClient
			logger.Info("employee names via gRPC", "employees", cfg.EmployeesGRPCAddr)
		}
	} else {
		logger.Warn("EMPLOYEES_GRPC_ADDR not set; movement document employee names disabled")
	}
	svc := service.NewPartService(repo, folderRepo, stockRepo, movementRepo, movementDocRepo, workOrders, brands, dealerPoints)

	gsrv := grpc.NewServer(observe.GRPCServerOptions(serviceName, logger, &grpcauth.Config{
		JWTSecret:  cfg.JWTSecret,
		WriteRoles: []string{"admin", "manager", "parts_manager", "storekeeper", "master", "service_advisor"},
	})...)
	partsv1.RegisterPartsServiceServer(gsrv, grpcserver.NewServer(svc, employees, dpChecker, woNotifier))
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
