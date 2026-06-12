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

	"github.com/dealer/dealer/pkg/dbschema"
	"github.com/dealer/dealer/pkg/health"
	"github.com/dealer/dealer/pkg/observe"
	"github.com/dealer/dealer/pkg/postgres"
	"github.com/dealer/dealer/scheduler-service/internal/config"
	"github.com/dealer/dealer/scheduler-service/internal/repository"
	"github.com/dealer/dealer/scheduler-service/internal/service"
	"github.com/dealer/dealer/scheduler-service/internal/worker"
)

const serviceName = "scheduler-service"

func main() {
	cfg := config.Load()
	logger := observe.Init(serviceName)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN,
		dbschema.Reviews, dbschema.WorkOrders, dbschema.Deals, dbschema.Customers,
		dbschema.Clients, dbschema.Vehicles, dbschema.Public,
	)
	if err != nil {
		logger.Error("postgres connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	invRepo := repository.NewInvitationRepository(pool)
	notifRepo := repository.NewNotificationRepository(pool)
	svc := service.NewSchedulerService(invRepo, notifRepo, cfg.BatchSize)
	w := worker.NewWorker(svc, cfg.PollInterval, logger)
	go w.Run(ctx)

	httpMux := http.NewServeMux()
	observe.RegisterHTTP(httpMux, health.Postgres(pool))
	httpHandler := observe.WrapHTTP(serviceName, httpMux, logger)
	httpLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.HTTPPort))
	if err != nil {
		logger.Error("http listen failed", "err", err)
		os.Exit(1)
	}
	go func() {
		logger.Info("HTTP listening", "port", cfg.HTTPPort, "poll_interval", cfg.PollInterval.String())
		if err := http.Serve(httpLis, httpHandler); err != nil {
			logger.Error("http serve failed", "err", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	stop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = shutdownCtx
}
