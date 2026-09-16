package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"commerce-platform/shared/database"

	"commerce-platform/services/orders/config"
	grpcx "commerce-platform/services/orders/internal/grpc"
	httpx "commerce-platform/services/orders/internal/http"
	"commerce-platform/services/orders/internal/repository"
	"commerce-platform/services/orders/internal/service"
	"commerce-platform/services/orders/migrations"
	loggerx "commerce-platform/shared/logger"
	shutdownx "commerce-platform/shared/shutdown"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
	httpPort := ":" + fmt.Sprint(cfg.Server.HTTPPort)
	productsGrpcClient := cfg.Products.GrpcClient
	shutdownTimeout := time.Duration(cfg.Server.GracefulShutdown.Timeout) * time.Second

	// import shared logger
	logger := loggerx.New(loggerx.Config{
		Service: "orders",
		Env:     cfg.Environment,
		Level:   loggerx.LevelFromEnv("LOG_LEVEL", zerolog.InfoLevel),
	})
	loggerx.SetAsDefault(logger)

	// ---- Database Health Check and Connection ----
	ctx := context.Background()
	db, err := database.NewPostgreClient(ctx, cfg.Database)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	if err := db.Ping(ctx); err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to postgres")
	}

	// ---- Migrations ----
	err = database.RunMigrations(cfg.Database, migrations.Files)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to run database migrations")
	}
	defer db.Close()

	logger.Info().Msg("Commerce Platform - ORDERS")
	r := chi.NewRouter()
	r.Use(loggerx.RequestContextMiddleware(logger))

	healthHandler := httpx.NewHealthHandler()
	healthHandler.RegisterRoutes(r)

	repo := repository.NewInMemoryOrderRepository()
	productsClient := grpcx.MustNewProductsGrpcClient(productsGrpcClient)
	svc := service.NewOrderService(repo, productsClient)
	orderHandler := httpx.NewOrderHandler(svc)
	orderHandler.RegisterRoutes(r)

	srv := &http.Server{Addr: httpPort, Handler: r}

	go func() {
		logger.Info().Msgf("HTTP server running on %s", httpPort)
		logger.Info().Msgf("Active Profile: %s", cfg.Profile)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("http server failed")
		}
	}()

	if err := shutdownx.WaitForTerminationAndShutdown(
		context.Background(), shutdownTimeout, func(shutdownCtx context.Context) error {
			logger.Info().Msg("shutdown signal received, draining connections")
			if err := srv.Shutdown(shutdownCtx); err != nil {
				logger.Error().Err(err).Msg("http server did not shut down cleanly")
				return err
			}

			logger.Info().Msg("http server stopped")
			return nil
		}); err != nil {
		logger.Error().Err(err).Msg("orders service shutdown failed")
	}
}
