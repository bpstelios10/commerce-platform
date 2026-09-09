package main

import (
	"context"
	"net"
	"net/http"

	grpcx "commerce-platform/services/products/internal/grpc"
	httpx "commerce-platform/services/products/internal/http"
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"
	loggerx "commerce-platform/shared/logger"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func main() {
	// import shared logger
	logger := loggerx.New(loggerx.Config{
		Service: "products",
		Env:     "local",
		Level:   zerolog.InfoLevel,
	})

	product1 := product.Product{
		ID:       uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d001"),
		Name:     "MacBook Pro",
		Category: "ACCESSORY",
		Price:    2500,
	}

	logger.Info().Msg(product1.DisplayName())

	product1.Rename("MacBook Pro M4")

	logger.Info().Msg(product1.Name)

	product1.ApplyDiscount(10)

	logger.Info().Msgf("product1: price=%v", product1.Price)

	logger.Info().Msgf("product1: is expensive=%v", product1.IsExpensive())

	logger.Info().Msg("--- TESTING CODE ---")
	products := map[string]product.Product{
		"1": {
			ID:       uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d002"),
			Name:     "MacBook Pro",
			Category: "ACCESSORY",
			Price:    2500,
		},
	}

	p, found := products["1"]
	logger.Info().Msgf("product found: product=%v, found=%v", p, found)

	p2, found2 := products["999"]
	logger.Info().Msgf("product found: product=%v, found=%v", p2, found2)

	// if i set the type, then i cant inject it to admin-service
	// var productRepo service.ProductRepository
	productRepo := repository.NewInMemoryProductRepository()

	logger.Info().Msgf("products loaded: %v", productRepo.FindAll(context.Background()))

	logger.Info().Msg("--- REAL LOGIC REST---")

	logger.Info().Msg("Commerce Platform - PRODUCTS")
	r := chi.NewRouter()
	r.Use(loggerx.RequestContextMiddleware(logger))

	productService := service.NewProductService(productRepo)
	productHandler := httpx.NewProductHandler(productService)
	productHandler.RegisterRoutes(r)

	healthHandler := httpx.NewHealthHandler()
	healthHandler.RegisterRoutes(r)

	categoryRepo := repository.NewInMemoryProductCategoryRepository()
	categoryService := service.NewProductCategoryService(categoryRepo)
	categoryHandler := httpx.NewProductCategoryHandler(categoryService)
	categoryHandler.RegisterRoutes(r)

	adminProductService := service.NewAdminService(productService, categoryService, productRepo)
	adminHandler := httpx.NewAdminHandler(adminProductService)
	adminHandler.RegisterRoutes(r)

	// this starts a go routine, like lightweight thread (in parallel).
	go func() {
		logger.Info().Msg("http server running on :8082")
		logger.Fatal().Err(http.ListenAndServe(":8082", r)).Msg("http server stopped")
	}()

	logger.Info().Msg("--- and gRPC ---")
	grpcHandler := grpcx.NewProductGrpcHandler(productService)
	grpcServer := grpc.NewServer()
	grpcx.RegisterProductServiceServer(
		grpcServer,
		grpcHandler,
	)

	// start gRPC
	lis, err := net.Listen("tcp", ":8092")
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to listen for grpc")
	}
	logger.Info().Msg("grpc server running on :8092")
	logger.Fatal().Err(grpcServer.Serve(lis)).Msg("grpc server stopped")
}
