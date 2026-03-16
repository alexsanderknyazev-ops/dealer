package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/dealer/dealer/services/vehicles/internal/config"
	grpcserver "github.com/dealer/dealer/services/vehicles/internal/grpc"
	"github.com/dealer/dealer/services/vehicles/internal/httpapi"
	"github.com/dealer/dealer/services/vehicles/internal/repository"
	"github.com/dealer/dealer/services/vehicles/internal/service"
	"github.com/dealer/dealer/pkg/postgres"
	vehiclesv1 "github.com/dealer/dealer/pkg/pb/vehicles/v1"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	repo := repository.NewVehicleRepository(pool)
	svc := service.NewVehicleService(repo)

	gsrv := grpc.NewServer()
	vehiclesv1.RegisterVehiclesServiceServer(gsrv, grpcserver.NewServer(svc))
	reflection.Register(gsrv)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}
	go func() {
		log.Printf("vehicles-service gRPC on :%d", cfg.GRPCPort)
		_ = gsrv.Serve(lis)
	}()

	httpMux := http.NewServeMux()
	httpapi.NewHandler(svc, cfg.JWTSecret).RegisterRoutes(httpMux)
	httpLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.HTTPPort))
	if err != nil {
		log.Fatalf("http listen: %v", err)
	}
	go func() {
		log.Printf("vehicles-service HTTP on :%d", cfg.HTTPPort)
		_ = http.Serve(httpLis, httpMux)
	}()

	<-ctx.Done()
	log.Println("shutting down...")
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
