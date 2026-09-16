package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"commerce-platform/services/products/config"
	grpcx "commerce-platform/services/products/internal/grpc"
	httpx "commerce-platform/services/products/internal/http"
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"
	"commerce-platform/shared/database"
	loggerx "commerce-platform/shared/logger"
	shutdownx "commerce-platform/shared/shutdown"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
	httpPort := ":" + fmt.Sprint(cfg.Server.HTTPPort)
	grpcPort := ":" + fmt.Sprint(cfg.Server.GRPCPort)
	shutdownTimeout := time.Duration(cfg.Server.GracefulShutdown.Timeout) * time.Second

	// import shared logger
	logger := loggerx.New(loggerx.Config{
		Service: "products",
		Env:     cfg.Environment,
		Level:   loggerx.LevelFromEnv("LOG_LEVEL", zerolog.InfoLevel),
	})
	loggerx.SetAsDefault(logger)

	// ---- Database Health Check and Connection ----
	ctx := context.Background()
	db, err := database.NewPostgreClient(ctx, cfg.Database)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	if err := db.Ping(ctx); err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer db.Close()

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

	httpServer := &http.Server{Addr: httpPort, Handler: r}

	logger.Info().Msg("--- and gRPC ---")
	grpcHandler := grpcx.NewProductGrpcHandler(productService)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpcx.LoggingUnaryInterceptor(logger)))
	grpcx.RegisterProductServiceServer(
		grpcServer,
		grpcHandler,
	)

	// start gRPC
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to listen for grpc")
	}

	go func() {
		logger.Info().Msgf("HTTP server running on %s", httpPort)
		logger.Info().Msgf("Active Profile: %s", cfg.Profile)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("http server failed")
		}
	}()

	go func() {
		logger.Info().Msgf("GRPC server running on %s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal().Err(err).Msg("grpc server failed")
		}
	}()

	if err := shutdownx.WaitForTerminationAndShutdown(
		context.Background(), shutdownTimeout, func(shutdownCtx context.Context) error {
			logger.Info().Msg("shutdown signal received, draining connections")
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				logger.Error().Err(err).Msg("http server did not shut down cleanly")
				return err
			}

			shutdownx.StopGracefullyOrForcefully(shutdownCtx, grpcServer.GracefulStop, grpcServer.Stop)
			logger.Info().Msg("products service stopped")
			return nil
		}); err != nil {
		logger.Error().Err(err).Msg("products service shutdown failed")
	}
}
