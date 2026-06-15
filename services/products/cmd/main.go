package main

import (
	"log"
	"log/slog"
	"net"
	"net/http"

	grpcx "commerce-platform/services/products/internal/grpc"
	httpx "commerce-platform/services/products/internal/http"
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"
	"commerce-platform/shared/logger"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func main() {
	// import shared logger
	logger := logger.New(logger.Config{
		Service: "products",
		Env:     "local",
		Level:   slog.LevelInfo,
	})
	slog.SetDefault(logger)

	product1 := product.Product{
		ID:    uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d001"),
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
			ID:    uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d002"),
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

	slog.Info("--- REAL LOGIC REST---")

	r := chi.NewRouter()

	productService := service.NewProductService(productRepo)
	productHandler := httpx.NewProductHandler(productService)
	productHandler.RegisterRoutes(r)

	healthHandler := httpx.NewHealthHandler()
	healthHandler.RegisterRoutes(r)

	adminProductService := service.NewAdminService(productService, productRepo)
	adminHandler := httpx.NewAdminHandler(adminProductService)
	adminHandler.RegisterRoutes(r)

	// this starts a go routine, like lightweight thread (in parallel).
	go func() {
		log.Println("http server running on :8080")
		http.ListenAndServe(":8080", r)
	}()

	slog.Info("--- and gRPC ---")
	grpcHandler := grpcx.NewProductGrpcHandler(productService)
	grpcServer := grpc.NewServer()
	grpcx.RegisterProductServiceServer(
		grpcServer,
		grpcHandler,
	)

	// start gRPC
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("grpc server running on :9090")
	grpcServer.Serve(lis)
}
