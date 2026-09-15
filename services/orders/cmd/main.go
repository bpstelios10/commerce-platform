package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"commerce-platform/services/orders/config"
	grpcx "commerce-platform/services/orders/internal/grpc"
	httpx "commerce-platform/services/orders/internal/http"
	"commerce-platform/services/orders/internal/repository"
	"commerce-platform/services/orders/internal/service"
	loggerx "commerce-platform/shared/logger"
	shutdownx "commerce-platform/shared/shutdown"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

const shutdownTimeout = 10 * time.Second

func main() {
	// import shared logger
	logger := loggerx.New(loggerx.Config{
		Service: "orders",
		Env:     "local",
		Level:   loggerx.LevelFromEnv("LOG_LEVEL", zerolog.InfoLevel),
	})
	loggerx.SetAsDefault(logger)

	logger.Info().Msg("Commerce Platform - ORDERS")
	r := chi.NewRouter()
	r.Use(loggerx.RequestContextMiddleware(logger))

	healthHandler := httpx.NewHealthHandler()
	healthHandler.RegisterRoutes(r)

	repo := repository.NewInMemoryOrderRepository()
	productsClient := grpcx.MustNewProductsGrpcClient("localhost:8092")
	svc := service.NewOrderService(repo, productsClient)
	orderHandler := httpx.NewOrderHandler(svc)
	orderHandler.RegisterRoutes(r)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load configuration")
	}
	httpPort := ":" + fmt.Sprint(cfg.Server.HTTPPort)

	srv := &http.Server{Addr: httpPort, Handler: r}

	go func() {
		logger.Info().Msg("http server running on " + httpPort)
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
