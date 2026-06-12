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
	appointmentsv1 "github.com/dealer/dealer/pkg/pb/appointments/v1"
	"github.com/dealer/dealer/pkg/postgres"
	"github.com/dealer/dealer/services/appointments/internal/client"
	"github.com/dealer/dealer/services/appointments/internal/config"
	grpcserver "github.com/dealer/dealer/services/appointments/internal/grpc"
	"github.com/dealer/dealer/services/appointments/internal/repository"
	"github.com/dealer/dealer/services/appointments/internal/service"
)

const serviceName = "appointments-service"

func main() {
	cfg := config.Load()
	logger := observe.Init(serviceName)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN, dbschema.Appointments, dbschema.Public)
	if err != nil {
		logger.Error("postgres connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.NewRepairAppointmentRepository(pool)

	var refs service.ReferenceChecker
	var displayRefs grpcserver.ReferenceDisplayer
	if cfg.CustomersGRPCAddr != "" && cfg.VehiclesGRPCAddr != "" {
		dialCtx, dialCancel := context.WithTimeout(ctx, 10*time.Second)
		refClient, err := client.NewReferenceChecker(
			dialCtx, cfg.CustomersGRPCAddr, cfg.VehiclesGRPCAddr, cfg.DealerPointsGRPCAddr,
			cfg.PartsGRPCAddr, cfg.WorksGRPCAddr,
		)
		dialCancel()
		if err != nil {
			logger.Warn("reference clients connect failed", "err", err)
		} else {
			defer refClient.Close()
			refs = refClient
			displayRefs = refClient
		}
	}

	var woClient *client.WorkOrdersCreator
	if cfg.WorkOrdersGRPCAddr != "" {
		dialCtx, dialCancel := context.WithTimeout(ctx, 10*time.Second)
		woClient, err = client.NewWorkOrdersCreator(dialCtx, cfg.WorkOrdersGRPCAddr)
		dialCancel()
		if err != nil {
			logger.Warn("workorders client connect failed", "err", err)
		} else if woClient != nil {
			defer woClient.Close()
		}
	}

	svc := service.NewRepairAppointmentService(repo, refs, woClient)

	gsrv := grpc.NewServer(observe.GRPCServerOptions(serviceName, logger, &grpcauth.Config{
		JWTSecret:  cfg.JWTSecret,
		WriteRoles: []string{"admin", "manager", "master", "service_advisor"},
	})...)
	appointmentsv1.RegisterRepairAppointmentsServiceServer(gsrv, grpcserver.NewServer(svc, displayRefs, woClient))
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
}
