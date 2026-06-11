package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dealer/dealer/pkg/observe"
	"github.com/dealer/dealer/services/gateway/client-protected/internal/config"
	"github.com/dealer/dealer/services/gateway/client-protected/internal/server"
)

const serviceName = "client-protected-gateway-service"

func main() {
	cfg := config.Load()
	logger := observe.Init(serviceName)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv, err := server.New(cfg)
	if err != nil {
		logger.Error("client protected gateway init failed", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	observe.RegisterHTTP(mux, nil)
	mux.Handle("/", srv.Handler())

	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: observe.WrapHTTP(serviceName, mux, logger),
	}

	go func() {
		logger.Info("client protected gateway listening", "addr", addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("client protected gateway serve failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down client protected gateway")
	stop()
	_ = httpSrv.Shutdown(context.Background())
}
