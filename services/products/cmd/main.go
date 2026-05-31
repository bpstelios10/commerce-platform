package main

import (
	"net/http"

	handler "commerce-platform/services/products/internal/http"
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"

	"fmt"

	"github.com/go-chi/chi/v5"
)

func main() {
	product1 := product.Product{
		ID:    "1",
		Name:  "MacBook Pro",
		Price: 2500,
	}

	fmt.Println(product1.DisplayName())

	product1.Rename("MacBook Pro M4")

	fmt.Println(product1.Name)

	product1.ApplyDiscount(10)

	fmt.Println(product1.Price)

	fmt.Println(product1.IsExpensive())

	fmt.Println("--- TESTING CODE ---")
	products := map[string]product.Product{
		"1": {
			ID:    "1",
			Name:  "MacBook Pro",
			Price: 2500,
		},
	}

	p, found := products["1"]
	fmt.Println(p)
	fmt.Println(found)

	p2, found2 := products["999"]
	fmt.Println(p2)
	fmt.Println(found2)

	var productRepo service.ProductRepository
	productRepo = repository.NewInMemoryProductRepository()

	fmt.Println(productRepo.FindAll())

	fmt.Println("--- REAL LOGIC ---")

	r := chi.NewRouter()

	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)
	productHandler.RegisterRoutes(r)

	healthHandler := handler.NewHealthHandler()
	healthHandler.RegisterRoutes(r)

	adminHandler := handler.NewAdminHandler()
	adminHandler.RegisterRoutes(r)

	http.ListenAndServe(":8080", r)
}
