package main

import (
	"fmt"
	"net/http"

	httpx "commerce-platform/services/orders/internal/http"
	"commerce-platform/services/orders/internal/repository"
	"commerce-platform/services/orders/internal/service"

	"github.com/go-chi/chi/v5"
)

func main() {
	fmt.Println("Commerce Platform - ORDERS")

	r := chi.NewRouter()

	healthHandler := httpx.NewHealthHandler()
	healthHandler.RegisterRoutes(r)

	repo := repository.NewInMemoryOrderRepository()
	svc := service.NewOrderService(repo)
	orderHandler := httpx.NewOrderHandler(svc)
	orderHandler.RegisterRoutes(r)

	http.ListenAndServe(":8080", r)
}
