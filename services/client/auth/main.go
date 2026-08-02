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

	"github.com/dealer/dealer/client-auth-service/internal/config"
	"github.com/dealer/dealer/client-auth-service/internal/consumer"
	clientauthgrpc "github.com/dealer/dealer/client-auth-service/internal/grpc"
	"github.com/dealer/dealer/client-auth-service/internal/repository"
	"github.com/dealer/dealer/client-auth-service/internal/service"
	"github.com/dealer/dealer/pkg/dbschema"
	"github.com/dealer/dealer/pkg/kafka"
	"github.com/dealer/dealer/pkg/observe"
	clientauthv1 "github.com/dealer/dealer/pkg/pb/clientauth/v1"
	"github.com/dealer/dealer/pkg/postgres"
	"github.com/dealer/dealer/pkg/redis"
)

const serviceName = "client-auth-service"

func main() {
	cfg := config.Load()
	logger := observe.Init(serviceName)
	tracerShutdown := observe.InitTracing(serviceName)
	defer tracerShutdown()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN, dbschema.ClientAuth, dbschema.Public)
	if err != nil {
		logger.Error("postgres connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	rdb := redis.NewClient(cfg.RedisAddr, cfg.RedisPass, cfg.RedisDB)
	if err := redis.Ping(ctx, rdb); err != nil {
		logger.Error("redis connect failed", "err", err)
		os.Exit(1)
	}
	defer rdb.Close()

	userRepo := repository.NewUserRepository(pool)
	authSvc := service.NewAuthService(userRepo, rdb, service.AuthConfig{
		JWTSecret:     cfg.JWTSecret,
		AccessTTL:     cfg.AccessTTL,
		RefreshTTL:    cfg.RefreshTTL,
		RefreshPrefix: cfg.RefreshPrefix,
	})

	if len(cfg.KafkaBrokers) > 0 && cfg.KafkaBrokers[0] != "" && cfg.KafkaTopic != "" {
		kConsumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaConsumerGroup)
		if kConsumer != nil {
			defer kConsumer.Close()
			go consumer.NewWorker(kConsumer, authSvc, logger).Run(ctx)
			logger.Info("kafka consumer started", "topic", cfg.KafkaTopic, "group", cfg.KafkaConsumerGroup)
		}
	} else {
		logger.Warn("kafka consumer disabled: KAFKA_BROKERS or KAFKA_TOPIC_CLIENT_REGISTRATION not set")
	}

	gsrv := grpc.NewServer(observe.GRPCServerOptions(serviceName, logger, nil)...)
	authSrv := clientauthgrpc.NewServer(authSvc)
	clientauthv1.RegisterClientAuthServiceServer(gsrv, authSrv)
	clientauthv1.RegisterClientAuthPublicServiceServer(gsrv, authSrv)
	clientauthv1.RegisterClientAuthSessionServiceServer(gsrv, authSrv)
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
	observe.RegisterHTTP(httpMux, func(ctx context.Context) error {
		if err := pool.Ping(ctx); err != nil {
			return err
		}
		return redis.Ping(ctx, rdb)
	})
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
