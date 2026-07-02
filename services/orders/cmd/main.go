package main

import (
	"log"
	"log/slog"
	"net/http"

	grpcx "commerce-platform/services/orders/internal/grpc"
	httpx "commerce-platform/services/orders/internal/http"
	"commerce-platform/services/orders/internal/repository"
	"commerce-platform/services/orders/internal/service"
	"commerce-platform/shared/logger"

	"github.com/go-chi/chi/v5"
)

func main() {
	// import shared logger
	logger := logger.New(logger.Config{
		Service: "orders",
		Env:     "local",
		Level:   slog.LevelInfo,
	})
	slog.SetDefault(logger)

	slog.Info("Commerce Platform - ORDERS")
	r := chi.NewRouter()

	healthHandler := httpx.NewHealthHandler()
	healthHandler.RegisterRoutes(r)

	repo := repository.NewInMemoryOrderRepository()
	productsClient := grpcx.MustNewProductsGrpcClient("localhost:8092")
	svc := service.NewOrderService(repo, productsClient)
	orderHandler := httpx.NewOrderHandler(svc)
	orderHandler.RegisterRoutes(r)

	log.Println("http server running on :8083")
	log.Fatal(http.ListenAndServe(":8083", r))
}
