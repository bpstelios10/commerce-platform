package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	grpcx "commerce-platform/services/orders/internal/grpc"
	httpx "commerce-platform/services/orders/internal/http"
	"commerce-platform/services/orders/internal/repository"
	"commerce-platform/services/orders/internal/service"
	loggerx "commerce-platform/shared/logger"

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

	srv := &http.Server{Addr: ":8083", Handler: r}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info().Msg("http server running on :8083")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("http server failed")
		}
	}()

	<-ctx.Done()
	stop()
	logger.Info().Msg("shutdown signal received, draining connections")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("http server did not shut down cleanly")
	}

	logger.Info().Msg("http server stopped")
}
