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

	"food_delivery_platform/services/auth-service/internal/application"
	"food_delivery_platform/services/auth-service/internal/config"
	httpdelivery "food_delivery_platform/services/auth-service/internal/delivery/http"
	"food_delivery_platform/services/auth-service/internal/domain/ports"
	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
	"food_delivery_platform/services/auth-service/internal/infrastructure/messaging"
	mongorepo "food_delivery_platform/services/auth-service/internal/infrastructure/mongo"
	"food_delivery_platform/services/auth-service/internal/infrastructure/security"
	"food_delivery_platform/services/auth-service/internal/infrastructure/system"
	"food_delivery_platform/services/auth-service/internal/observability"
	"food_delivery_platform/shared/broker"
	"food_delivery_platform/shared/logger"

	mdriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	cfg := config.Load()
	log := logger.New(logger.Config{ServiceName: cfg.ServiceName, Level: cfg.LogLevel})

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startupCancel()

	tracingShutdown, err := observability.SetupTracing(startupCtx, observability.TracingConfig{
		Enabled:        cfg.OTelEnabled,
		ServiceName:    cfg.ServiceName,
		ServiceVersion: cfg.ServiceVersion,
		Environment:    os.Getenv("ENVIRONMENT"),
	})
	if err != nil {
		log.Error("failed to initialize tracing", "error", err.Error())
		os.Exit(1)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tracingShutdown(ctx); err != nil {
			log.Error("failed to shutdown tracing", "error", err.Error())
		}
	}()

	client, db, err := connectMongo(startupCtx, cfg.MongoURI, cfg.MongoDBName, cfg.OTelEnabled)
	if err != nil {
		log.Error("failed to connect mongo", "error", err.Error())
		os.Exit(1)
	}

	if err := mongorepo.EnsureIndexes(startupCtx, db); err != nil {
		log.Error("failed to ensure mongo indexes", "error", err.Error())
		os.Exit(1)
	}

	credRepo := mongorepo.NewCredentialRepository(db)
	sessionRepo := mongorepo.NewRefreshSessionRepository(db)

	hasher, err := security.NewBcryptPasswordHasher(cfg.BcryptCost)
	if err != nil {
		log.Error("failed to initialize password hasher", "error", err.Error())
		os.Exit(1)
	}

	tokenService, err := security.NewHMACTokenService(cfg.JWTSecret, cfg.JWTIssuer)
	if err != nil {
		log.Error("failed to initialize token service", "error", err.Error())
		os.Exit(1)
	}

	ttl, err := valueobjects.NewTokenTTL(cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	if err != nil {
		log.Error("failed to initialize token ttl", "error", err.Error())
		os.Exit(1)
	}

	clock := system.Clock{}
	idGen := system.IDGenerator{}

	var eventPublisher ports.AuthEventPublisher
	var rabbit *broker.Rabbit
	outboxRepo := mongorepo.NewOutboxRepository(db)
	if cfg.RabbitMQURL != "" {
		rabbit, err = broker.New(broker.Config{URL: cfg.RabbitMQURL})
		if err != nil {
			log.Error("failed to connect rabbitmq", "error", err.Error())
			os.Exit(1)
		}

		ch := rabbit.Channel()
		if err := ch.ExchangeDeclare("auth.exchange", "topic", true, false, false, false, nil); err != nil {
			log.Error("failed to declare auth exchange", "error", err.Error())
			_ = rabbit.Close()
			os.Exit(1)
		}

		eventPublisher = messaging.NewRabbitAuthEventPublisher(
			ch,
			"auth.exchange",
			cfg.ServiceName,
			log,
			broker.RetryPolicy{MaxAttempts: 3, BaseDelay: 200 * time.Millisecond, MaxDelay: 2 * time.Second},
			outboxRepo,
		)
	}

	registerUC := application.NewRegisterUserUseCase(credRepo, sessionRepo, hasher, tokenService, clock, idGen, ttl, eventPublisher)
	loginUC := application.NewLoginUserUseCase(credRepo, sessionRepo, hasher, tokenService, clock, idGen, ttl, eventPublisher)
	refreshUC := application.NewRefreshAccessTokenUseCase(credRepo, sessionRepo, tokenService, clock, idGen, ttl)
	logoutUC := application.NewLogoutSessionUseCase(sessionRepo, tokenService)
	authHandlers := httpdelivery.NewAuthHandlers(registerUC, loginUC, refreshUC, logoutUC)

	handler := httpdelivery.NewRouter(log, cfg.RequestTimeout, authHandlers)
	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("auth-service starting", "port", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", "error", err.Error())
			os.Exit(1)
		}
	}()

	waitForShutdown(log, server, client, rabbit)
}

func waitForShutdown(log *slog.Logger, server *http.Server, mongoClient *mdriver.Client, rabbit *broker.Rabbit) {
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
	if mongoClient != nil {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Error("mongo disconnect failed", "error", err.Error())
		}
	}
	if rabbit != nil {
		if err := rabbit.Close(); err != nil {
			log.Error("rabbit disconnect failed", "error", err.Error())
		}
	}
	log.Info("server stopped")
}

func connectMongo(ctx context.Context, uri, dbName string, withTelemetry bool) (*mdriver.Client, *mdriver.Database, error) {
	if uri == "" {
		return nil, nil, errors.New("MONGO_URI is required")
	}
	if dbName == "" {
		return nil, nil, errors.New("MONGO_DB_NAME is required")
	}

	clientOpts := options.Client().ApplyURI(uri)
	if withTelemetry {
		clientOpts.SetMonitor(observability.NewMongoMonitor())
	}

	client, err := mdriver.Connect(ctx, clientOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("mongo connect: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, fmt.Errorf("mongo ping: %w", err)
	}

	return client, client.Database(dbName), nil
}
