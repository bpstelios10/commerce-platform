package main

import (
	"log"
	"net/http"

	grpcx "commerce-platform/services/orders/internal/grpc"
	httpx "commerce-platform/services/orders/internal/http"
	"commerce-platform/services/orders/internal/repository"
	"commerce-platform/services/orders/internal/service"

	"github.com/go-chi/chi/v5"
)

func main() {
	log.Println("Commerce Platform - ORDERS")
	r := chi.NewRouter()

	healthHandler := httpx.NewHealthHandler()
	healthHandler.RegisterRoutes(r)

	repo := repository.NewInMemoryOrderRepository()
	productsClient := grpcx.MustNewProductsGrpcClient("localhost:9090")
	svc := service.NewOrderService(repo, productsClient)
	orderHandler := httpx.NewOrderHandler(svc)
	orderHandler.RegisterRoutes(r)

	log.Println("http server running on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}
