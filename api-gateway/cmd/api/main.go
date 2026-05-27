package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"food_delivery_platform/api-gateway/internal/config"
	httpdelivery "food_delivery_platform/api-gateway/internal/delivery/http"
	"food_delivery_platform/api-gateway/internal/proxy"
	"food_delivery_platform/shared/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(logger.Config{
		ServiceName: cfg.ServiceName,
		Level:       cfg.LogLevel,
	})

	authProxy, err := proxy.NewAuthProxy(cfg.AuthServiceURL, log)
	if err != nil {
		log.Error("invalid AUTH_SERVICE_URL", "error", err.Error())
		os.Exit(1)
	}

	handler := httpdelivery.NewRouter(log, cfg.RequestTimeout, authProxy)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("api-gateway starting", "port", cfg.HTTPPort, "auth_upstream", cfg.AuthServiceURL)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "error", err.Error())
			os.Exit(1)
		}
	}()

	waitForShutdown(log, server)
}

func waitForShutdown(log *slog.Logger, server *http.Server) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info("shutdown signal received")
	if err := server.Shutdown(ctx); err != nil {
		log.Error("shutdown failed", "error", err.Error())
		return
	}
	log.Info("server stopped")
}
