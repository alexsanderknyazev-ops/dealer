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

	"github.com/dealer/dealer/client-reviews-service/internal/client"
	"github.com/dealer/dealer/client-reviews-service/internal/config"
	grpcserver "github.com/dealer/dealer/client-reviews-service/internal/grpc"
	"github.com/dealer/dealer/client-reviews-service/internal/publisher"
	"github.com/dealer/dealer/client-reviews-service/internal/repository"
	"github.com/dealer/dealer/client-reviews-service/internal/service"
	"github.com/dealer/dealer/pkg/dbschema"
	"github.com/dealer/dealer/pkg/health"
	"github.com/dealer/dealer/pkg/kafka"
	"github.com/dealer/dealer/pkg/observe"
	reviewsv1 "github.com/dealer/dealer/pkg/pb/reviews/v1"
	"github.com/dealer/dealer/pkg/postgres"
)

const serviceName = "client-reviews-service"

func main() {
	cfg := config.Load()
	logger := observe.Init(serviceName)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN, dbschema.Reviews, dbschema.Clients, dbschema.Public)
	if err != nil {
		logger.Error("postgres connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	dialCtx, dialCancel := context.WithTimeout(ctx, 10*time.Second)
	vehiclesClient, err := client.NewVehiclesClient(dialCtx, cfg.VehiclesGRPCAddr)
	dialCancel()
	if err != nil {
		logger.Error("vehicles gRPC connect failed", "err", err)
		os.Exit(1)
	}
	defer vehiclesClient.Close()

	repo := repository.NewReviewRepository(pool)
	var reviewPub *publisher.ReviewPublished
	if len(cfg.KafkaBrokers) > 0 && cfg.KafkaBrokers[0] != "" && cfg.KafkaTopic != "" {
		kp := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
		defer kp.Close()
		reviewPub = publisher.NewReviewPublished(kp)
		logger.Info("kafka review.published publisher enabled", "topic", cfg.KafkaTopic)
	} else {
		logger.Warn("kafka review.published publisher disabled")
	}
	svc := service.NewReviewService(repo, vehiclesClient, reviewPub)

	gsrv := grpc.NewServer(observe.GRPCServerOptions(serviceName, logger, nil)...)
	reviewsv1.RegisterReviewsServiceServer(gsrv, grpcserver.NewServer(svc, cfg.JWTSecret))
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
