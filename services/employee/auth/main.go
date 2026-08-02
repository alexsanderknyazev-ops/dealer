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

	"github.com/dealer/dealer/auth-service/internal/config"
	authgrpc "github.com/dealer/dealer/auth-service/internal/grpc"
	"github.com/dealer/dealer/auth-service/internal/httpapi"
	"github.com/dealer/dealer/auth-service/internal/repository"
	"github.com/dealer/dealer/auth-service/internal/service"
	"github.com/dealer/dealer/pkg/dbschema"
	"github.com/dealer/dealer/pkg/kafka"
	"github.com/dealer/dealer/pkg/observe"
	authv1 "github.com/dealer/dealer/pkg/pb/auth/v1"
	"github.com/dealer/dealer/pkg/postgres"
	"github.com/dealer/dealer/pkg/redis"
)

const serviceName = "auth-service"

func main() {
	cfg := config.Load()
	logger := observe.Init(serviceName)
	tracerShutdown := observe.InitTracing(serviceName)
	defer tracerShutdown()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN, dbschema.Auth, dbschema.Public)
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

	var publisher service.EventPublisher
	if len(cfg.KafkaBrokers) > 0 && cfg.KafkaBrokers[0] != "" {
		kp := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
		defer kp.Close()
		publisher = kp
	}

	userRepo := repository.NewUserRepository(pool)
	authSvc := service.NewAuthService(userRepo, rdb, publisher, service.AuthConfig{
		JWTSecret:     cfg.JWTSecret,
		AccessTTL:     cfg.AccessTTL,
		RefreshTTL:    cfg.RefreshTTL,
		RefreshPrefix: "auth:refresh:",
	})

	gsrv := grpc.NewServer(observe.GRPCServerOptions(serviceName, logger, nil)...)
	authv1.RegisterAuthServiceServer(gsrv, authgrpc.NewServer(authSvc))
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
	registerDomainAPIProxy(httpMux, cfg, logger)
	if cfg.StaticDir != "" {
		httpMux.Handle("/", httpapi.SPAFileServer(http.Dir(cfg.StaticDir)))
		logger.Info("serving static files", "dir", cfg.StaticDir)
	}
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
