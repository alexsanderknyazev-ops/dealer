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
	"github.com/dealer/dealer/pkg/kafka"
	"github.com/dealer/dealer/pkg/observe"
	"github.com/dealer/dealer/pkg/postgres"
	"github.com/dealer/dealer/scheduler-service/internal/config"
	"github.com/dealer/dealer/scheduler-service/internal/consumer"
	"github.com/dealer/dealer/scheduler-service/internal/repository"
	"github.com/dealer/dealer/scheduler-service/internal/service"
	"github.com/dealer/dealer/scheduler-service/internal/worker"
)

const serviceName = "scheduler-service"

func main() {
	cfg := config.Load()
	logger := observe.Init(serviceName)
	tracerShutdown := observe.InitTracing(serviceName)
	defer tracerShutdown()
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

	if len(cfg.KafkaBrokers) > 0 && cfg.KafkaBrokers[0] != "" && cfg.KafkaTopic != "" {
		kConsumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaConsumerGroup)
		if kConsumer != nil {
			defer kConsumer.Close()
			go consumer.NewWorker(kConsumer, notifRepo, logger).Run(ctx)
			logger.Info("kafka consumer started", "topic", cfg.KafkaTopic, "group", cfg.KafkaConsumerGroup)
		}
	} else {
		logger.Warn("kafka consumer disabled: KAFKA_BROKERS or KAFKA_TOPIC_APPOINTMENT_CREATED not set")
	}

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
