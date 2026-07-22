package main

import (
	"log/slog"
	"net/http"

	grpcx "commerce-platform/services/orders/internal/grpc"
	httpx "commerce-platform/services/orders/internal/http"
	"commerce-platform/services/orders/internal/repository"
	"commerce-platform/services/orders/internal/service"
	loggerx "commerce-platform/shared/logger"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func main() {
	// import shared logger
	logger := loggerx.New(loggerx.Config{
		Service: "orders",
		Env:     "local",
		Level:   zerolog.InfoLevel,
	})
	// set the default slog to point to logger, just in case
	slogHandler := zerolog.NewSlogHandler(logger)
	slog.SetDefault(slog.New(slogHandler))

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

	logger.Info().Msg("http server running on :8083")
	logger.Fatal().Err(http.ListenAndServe(":8083", r)).Msg("http server stopped")
}
