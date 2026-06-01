package main

import (
	"log/slog"
	"net/http"

	httpx "commerce-platform/services/products/internal/http"
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"

	"github.com/go-chi/chi/v5"
)

func main() {
	product1 := product.Product{
		ID:    "1",
		Name:  "MacBook Pro",
		Price: 2500,
	}

	slog.Info(product1.DisplayName())

	product1.Rename("MacBook Pro M4")

	slog.Info(product1.Name)

	product1.ApplyDiscount(10)

	slog.Info("product1", "price", product1.Price)

	slog.Info("product1", "is expensive", product1.IsExpensive())

	slog.Info("--- TESTING CODE ---")
	products := map[string]product.Product{
		"1": {
			ID:    "1",
			Name:  "MacBook Pro",
			Price: 2500,
		},
	}

	p, found := products["1"]
	slog.Info("product found", "product", p, "found", found)

	p2, found2 := products["999"]
	slog.Info("product found", "product", p2, "found", found2)

	// if i set the type, then i cant inject it to admin-service
	// var productRepo service.ProductRepository
	productRepo := repository.NewInMemoryProductRepository()

	slog.Info("products loaded", "products", productRepo.FindAll())

	slog.Info("--- REAL LOGIC ---")

	r := chi.NewRouter()

	productService := service.NewProductService(productRepo)
	productHandler := httpx.NewProductHandler(productService)
	productHandler.RegisterRoutes(r)

	healthHandler := httpx.NewHealthHandler()
	healthHandler.RegisterRoutes(r)

	adminProductService := service.NewAdminService(productRepo)
	adminHandler := httpx.NewAdminHandler(adminProductService)
	adminHandler.RegisterRoutes(r)

	http.ListenAndServe(":8080", r)
}
