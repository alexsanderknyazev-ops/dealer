package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
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
	"github.com/dealer/dealer/pkg/kafka"
	"github.com/dealer/dealer/pkg/postgres"
	"github.com/dealer/dealer/pkg/redis"
	authv1 "github.com/dealer/dealer/pkg/pb/auth/v1"
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

	rdb := redis.NewClient(cfg.RedisAddr, cfg.RedisPass, cfg.RedisDB)
	if err := redis.Ping(ctx, rdb); err != nil {
		log.Fatalf("redis: %v", err)
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

	gsrv := grpc.NewServer()
	authv1.RegisterAuthServiceServer(gsrv, authgrpc.NewServer(authSvc))
	reflection.Register(gsrv)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	go func() {
		log.Printf("auth-service gRPC listening on :%d", cfg.GRPCPort)
		if err := gsrv.Serve(lis); err != nil {
			log.Printf("grpc serve: %v", err)
		}
	}()

	// HTTP API + SPA для браузера
	httpMux := http.NewServeMux()
	httpapi.NewHandler(authSvc).RegisterRoutes(httpMux)
	if cfg.CustomersServiceURL != "" {
		if targetURL, err := url.Parse(cfg.CustomersServiceURL); err == nil {
			proxy := httputil.NewSingleHostReverseProxy(targetURL)
			httpMux.Handle("/api/customers", proxy)
			httpMux.Handle("/api/customers/", proxy)
			log.Printf("auth-service proxying /api/customers to %s", cfg.CustomersServiceURL)
		}
	}
	if cfg.VehiclesServiceURL != "" {
		if targetURL, err := url.Parse(cfg.VehiclesServiceURL); err == nil {
			proxy := httputil.NewSingleHostReverseProxy(targetURL)
			httpMux.Handle("/api/vehicles", proxy)
			httpMux.Handle("/api/vehicles/", proxy)
			log.Printf("auth-service proxying /api/vehicles to %s", cfg.VehiclesServiceURL)
		}
	}
	if cfg.DealsServiceURL != "" {
		if targetURL, err := url.Parse(cfg.DealsServiceURL); err == nil {
			proxy := httputil.NewSingleHostReverseProxy(targetURL)
			httpMux.Handle("/api/deals", proxy)
			httpMux.Handle("/api/deals/", proxy)
			log.Printf("auth-service proxying /api/deals to %s", cfg.DealsServiceURL)
		}
	}
	if cfg.PartsServiceURL != "" {
		if targetURL, err := url.Parse(cfg.PartsServiceURL); err == nil {
			proxy := httputil.NewSingleHostReverseProxy(targetURL)
			httpMux.Handle("/api/parts", proxy)
			httpMux.Handle("/api/parts/", proxy)
			log.Printf("auth-service proxying /api/parts to %s", cfg.PartsServiceURL)
		}
	}
	if cfg.BrandsServiceURL != "" {
		if targetURL, err := url.Parse(cfg.BrandsServiceURL); err == nil {
			proxy := httputil.NewSingleHostReverseProxy(targetURL)
			httpMux.Handle("/api/brands", proxy)
			httpMux.Handle("/api/brands/", proxy)
			log.Printf("auth-service proxying /api/brands to %s", cfg.BrandsServiceURL)
		}
	}
	if cfg.DealerPointsServiceURL != "" {
		if targetURL, err := url.Parse(cfg.DealerPointsServiceURL); err == nil {
			proxy := httputil.NewSingleHostReverseProxy(targetURL)
			httpMux.Handle("/api/dealer-points", proxy)
			httpMux.Handle("/api/dealer-points/", proxy)
			httpMux.Handle("/api/legal-entities", proxy)
			httpMux.Handle("/api/legal-entities/", proxy)
			httpMux.Handle("/api/warehouses", proxy)
			httpMux.Handle("/api/warehouses/", proxy)
			log.Printf("auth-service proxying /api/dealer-points, /api/legal-entities, /api/warehouses to %s", cfg.DealerPointsServiceURL)
		}
	}
	if cfg.StaticDir != "" {
		httpMux.Handle("/", httpapi.SPAFileServer(http.Dir(cfg.StaticDir)))
		log.Printf("auth-service serving static from %s", cfg.StaticDir)
	}
	httpLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.HTTPPort))
	if err != nil {
		log.Fatalf("http listen: %v", err)
	}
	go func() {
		log.Printf("auth-service HTTP listening on :%d", cfg.HTTPPort)
		if err := http.Serve(httpLis, httpMux); err != nil {
			log.Printf("http serve: %v", err)
		}
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
