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

	"github.com/dealer/dealer/pkg/kafka"
	"github.com/dealer/dealer/pkg/observe"
	"github.com/dealer/dealer/services/errors-ingest/internal/clickhouse"
	"github.com/dealer/dealer/services/errors-ingest/internal/config"
	"github.com/dealer/dealer/services/errors-ingest/internal/httpapi"
	"github.com/dealer/dealer/services/errors-ingest/internal/ingest"
)

const serviceName = "errors-ingest-service"

func main() {
	cfg := config.Load()
	logger := observe.Init(serviceName)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	store, err := clickhouse.Open(initCtx, cfg.ClickHouseAddr, cfg.ClickHouseDatabase, cfg.ClickHouseUser, cfg.ClickHousePassword)
	cancel()
	if err != nil {
		logger.Error("clickhouse connect failed", "err", err)
		os.Exit(1)
	}
	defer store.Close()

	consumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopicErrors, cfg.KafkaConsumerGroup)
	if consumer == nil {
		logger.Error("kafka consumer not configured")
		os.Exit(1)
	}
	defer consumer.Close()

	worker := ingest.NewWorker(consumer, store, logger)
	go worker.Run(ctx)

	httpMux := http.NewServeMux()
	httpapi.NewHandler(store, cfg.Environment).RegisterRoutes(httpMux)
	observe.RegisterHTTP(httpMux, func(c context.Context) error {
		return store.EnsureSchema(c)
	})
	httpHandler := observe.WrapHTTP(serviceName, httpMux, logger)

	httpLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.HTTPPort))
	if err != nil {
		logger.Error("http listen failed", "err", err)
		os.Exit(1)
	}
	go func() {
		logger.Info("errors-ingest listening", "port", cfg.HTTPPort, "clickhouse", cfg.ClickHouseAddr, "kafka_topic", cfg.KafkaTopicErrors)
		if err := http.Serve(httpLis, httpHandler); err != nil {
			logger.Error("http serve failed", "err", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	stop()
}
