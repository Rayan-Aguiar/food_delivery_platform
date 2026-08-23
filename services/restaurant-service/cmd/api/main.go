package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"food_delivery_platform/services/restaurant-service/internal/config"
	httpdelivery "food_delivery_platform/services/restaurant-service/internal/delivery/http"
	mongorepo "food_delivery_platform/services/restaurant-service/internal/infrastructure/mongo"
	"food_delivery_platform/shared/logger"

	mdriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	cfg := config.Load()
	log := logger.New(logger.Config{
		ServiceName: cfg.ServiceName,
		Level:       cfg.LogLevel,
	})

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startupCancel()

	mongoClient, db, err := connectMongo(startupCtx, cfg.MongoURI, cfg.MongoDBName)
	if err != nil {
		log.Error("failed to connect mongo", "error", err.Error())
		os.Exit(1)
	}

	if err := mongorepo.EnsureIndexes(startupCtx, db); err != nil {
		log.Error("failed to ensure mongo indexes", "error", err.Error())
		os.Exit(1)
	}

	handler := httpdelivery.NewRouter(log, cfg.RequestTimeout, cfg.ServiceName)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("restaurante-service starting", "port", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	waitForShutdown(log, server, mongoClient)

}

func waitForShutdown(log *slog.Logger, server *http.Server, mongoClient *mdriver.Client) {
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

	if err := mongoClient.Disconnect(ctx); err != nil {
		log.Error("mongo disconnect failed", "error", err.Error())
	}

	log.Info("server stopped")
}

func connectMongo(ctx context.Context, uri, dbName string) (*mdriver.Client, *mdriver.Database, error) {
	if uri == "" {
		return nil, nil, errors.New("MONGO_URI is required")
	}
	if dbName == "" {
		return nil, nil, errors.New("MONGO_DB_NAME is required")
	}

	client, err := mdriver.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, fmt.Errorf("mongo connect: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, fmt.Errorf("mongo ping: %w", err)
	}

	return client, client.Database(dbName), nil
}
