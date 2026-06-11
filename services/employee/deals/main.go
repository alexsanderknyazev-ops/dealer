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
	"github.com/dealer/dealer/pkg/kafka"
	"github.com/dealer/dealer/pkg/observe"
	dealsv1 "github.com/dealer/dealer/pkg/pb/deals/v1"
	"github.com/dealer/dealer/pkg/postgres"
	"github.com/dealer/dealer/services/deals/internal/client"
	"github.com/dealer/dealer/services/deals/internal/config"
	grpcserver "github.com/dealer/dealer/services/deals/internal/grpc"
	"github.com/dealer/dealer/services/deals/internal/publisher"
	"github.com/dealer/dealer/services/deals/internal/repository"
	"github.com/dealer/dealer/services/deals/internal/service"
)

const serviceName = "deals-service"

func main() {
	cfg := config.Load()
	logger := observe.Init(serviceName)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN, dbschema.Deals, dbschema.Public)
	if err != nil {
		logger.Error("postgres connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.NewDealRepository(pool)
	var refs service.ReferenceChecker
	if cfg.CustomersGRPCAddr != "" && cfg.VehiclesGRPCAddr != "" {
		dialCtx, dialCancel := context.WithTimeout(ctx, 10*time.Second)
		refClient, err := client.NewReferenceChecker(dialCtx, cfg.CustomersGRPCAddr, cfg.VehiclesGRPCAddr)
		dialCancel()
		if err != nil {
			logger.Error("gRPC ref clients connect failed", "err", err)
			os.Exit(1)
		}
		defer refClient.Close()
		refs = refClient
		logger.Info("reference checks via gRPC", "customers", cfg.CustomersGRPCAddr, "vehicles", cfg.VehiclesGRPCAddr)
	} else {
		logger.Warn("CUSTOMERS_GRPC_ADDR/VEHICLES_GRPC_ADDR not set; deal reference checks disabled")
	}
	var dealPublisher *publisher.DealCompleted
	if len(cfg.KafkaBrokers) > 0 && cfg.KafkaBrokers[0] != "" && cfg.KafkaTopic != "" {
		kp := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
		defer kp.Close()
		dealPublisher = publisher.NewDealCompleted(kp)
		logger.Info("kafka deal.completed publisher enabled", "topic", cfg.KafkaTopic)
	} else {
		logger.Warn("kafka deal.completed publisher disabled")
	}
	svc := service.NewDealService(repo, refs, dealPublisher)

	gsrv := grpc.NewServer(observe.GRPCServerOptions(serviceName, logger, &grpcauth.Config{
		JWTSecret:  cfg.JWTSecret,
		WriteRoles: []string{"admin", "manager", "sales"},
	})...)
	dealsv1.RegisterDealsServiceServer(gsrv, grpcserver.NewServer(svc))
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
