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

	"github.com/dealer/dealer/employee-statistics-service/internal/config"
	"github.com/dealer/dealer/employee-statistics-service/internal/consumer"
	grpcserver "github.com/dealer/dealer/employee-statistics-service/internal/grpc"
	"github.com/dealer/dealer/employee-statistics-service/internal/repository"
	"github.com/dealer/dealer/employee-statistics-service/internal/service"
	"github.com/dealer/dealer/pkg/dbschema"
	"github.com/dealer/dealer/pkg/grpcauth"
	"github.com/dealer/dealer/pkg/health"
	"github.com/dealer/dealer/pkg/kafka"
	"github.com/dealer/dealer/pkg/observe"
	employeestatsv1 "github.com/dealer/dealer/pkg/pb/statistics/employee/v1"
	"github.com/dealer/dealer/pkg/postgres"
)

const serviceName = "employee-statistics-service"

func main() {
	cfg := config.Load()
	logger := observe.Init(serviceName)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN, dbschema.EmployeeStatistics, dbschema.Public)
	if err != nil {
		logger.Error("postgres connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.NewStatsRepository(pool)
	svc := service.NewStatsService(repo)

	if len(cfg.KafkaBrokers) > 0 && cfg.KafkaBrokers[0] != "" && cfg.KafkaDealTopic != "" {
		kConsumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaDealTopic, cfg.KafkaConsumerGroup)
		if kConsumer != nil {
			defer kConsumer.Close()
			go consumer.NewDealWorker(kConsumer, svc, logger).Run(ctx)
			logger.Info("kafka consumer started", "topic", cfg.KafkaDealTopic, "group", cfg.KafkaConsumerGroup)
		}
	} else {
		logger.Warn("kafka consumer disabled for employee statistics")
	}

	gsrv := grpc.NewServer(observe.GRPCServerOptions(serviceName, logger, &grpcauth.Config{
		JWTSecret:  cfg.JWTSecret,
		WriteRoles: []string{"admin", "manager", "sales"},
	})...)
	employeestatsv1.RegisterEmployeeStatisticsServiceServer(gsrv, grpcserver.NewServer(svc))
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
