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
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"
	"commerce-platform/services/products/migrations"
	"commerce-platform/shared/database"
	loggerx "commerce-platform/shared/logger"
	shutdownx "commerce-platform/shared/shutdown"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func main() {
	// ---- Load configuration ----
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
	httpPort := ":" + fmt.Sprint(cfg.Server.HTTPPort)
	grpcPort := ":" + fmt.Sprint(cfg.Server.GRPCPort)
	shutdownTimeout := time.Duration(cfg.Server.GracefulShutdown.Timeout) * time.Second
	logLevel := cfg.GetLogLevel(zerolog.InfoLevel)

	// ---- Initialize shared logger ----
	logger := loggerx.New(loggerx.Config{
		Service: "products",
		Env:     cfg.Environment,
		Level:   logLevel,
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

	// ---- Migrations ----
	err = database.RunMigrations(cfg.Database, migrations.Files)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to run database migrations")
	}
	defer db.Close()

	// ---- Setup Server ----
	logger.Info().Msg("Commerce Platform - PRODUCTS")
	r := chi.NewRouter()
	r.Use(loggerx.RequestContextMiddleware(logger))

	// product handler
	productRepo := repository.NewPostgreProductRepository(db)
	productService := service.NewProductService(productRepo)
	productHandler := httpx.NewProductHandler(productService)
	productHandler.RegisterRoutes(r)

	// health handler
	healthHandler := httpx.NewHealthHandler()
	healthHandler.RegisterRoutes(r)

	// product category handler
	categoryRepo := repository.NewPostgreProductCategoryRepository(db)
	categoryService := service.NewProductCategoryService(categoryRepo)
	categoryHandler := httpx.NewProductCategoryHandler(categoryService)
	categoryHandler.RegisterRoutes(r)

	// admin handler
	adminProductService := service.NewAdminService(productService, categoryService, productRepo)
	adminHandler := httpx.NewAdminHandler(adminProductService)
	adminHandler.RegisterRoutes(r)

	httpServer := &http.Server{Addr: httpPort, Handler: r}

	// ---- Setup gRPC Server ----
	logger.Info().Msg("--- and gRPC ---")
	grpcHandler := grpcx.NewProductGrpcHandler(productService)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpcx.LoggingUnaryInterceptor(logger)))
	grpcx.RegisterProductServiceServer(
		grpcServer,
		grpcHandler,
	)

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to listen for grpc")
	}

	// ---- Start Servers with Graceful Shutdown ----
	go func() {
		logger.Info().Msgf("HTTP server running on %s", httpPort)
		logger.Info().Msgf("Active Profile: %s", cfg.Profile)
		logger.Info().Msgf("Log Level: %s", logLevel)

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
