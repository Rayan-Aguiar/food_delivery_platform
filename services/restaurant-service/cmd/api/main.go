package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"food_delivery_platform/services/restaurant-service/internal/config"
	httpdelivery "food_delivery_platform/services/restaurant-service/internal/delivery/http"
	"food_delivery_platform/shared/logger"
)


func main() {
	cfg := config.Load()
	log := logger.New(logger.Config{
		ServiceName: cfg.ServiceName,
		Level: cfg.LogLevel,
	})

	handler := httpdelivery.NewRouter(log, cfg.RequestTimeout, cfg.ServiceName)

	server := &http.Server{
		Addr: 	   ":" + cfg.HTTPPort,
		Handler:    handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("restaurante-service starting", "port", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start server", "error", err)
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