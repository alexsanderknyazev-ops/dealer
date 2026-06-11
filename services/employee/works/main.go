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
	worksv1 "github.com/dealer/dealer/pkg/pb/works/v1"
	"github.com/dealer/dealer/pkg/postgres"
	"github.com/dealer/dealer/services/works/internal/config"
	grpcserver "github.com/dealer/dealer/services/works/internal/grpc"
	"github.com/dealer/dealer/services/works/internal/repository"
	"github.com/dealer/dealer/services/works/internal/service"
)

const serviceName = "works-service"

func main() {
	cfg := config.Load()
	logger := observe.Init(serviceName)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN, dbschema.Works, dbschema.Public)
	if err != nil {
		logger.Error("postgres connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.NewWorkRepository(pool)
	svc := service.NewWorkService(repo)

	gsrv := grpc.NewServer(observe.GRPCServerOptions(serviceName, logger, &grpcauth.Config{
		JWTSecret:  cfg.JWTSecret,
		WriteRoles: []string{"admin", "manager", "master", "service_advisor"},
	})...)
	worksv1.RegisterWorksServiceServer(gsrv, grpcserver.NewServer(svc))
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
