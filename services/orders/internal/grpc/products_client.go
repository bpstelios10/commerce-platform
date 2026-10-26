package grpc

import (
	"context"
	"log"
	"time"

	"commerce-platform/shared/logger"

	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// ProductsGrpcClient adapts the generated ProductServiceClient to the
// service.ProductsClient interface, hiding gRPC boilerplate from the service layer.
type ProductsGrpcClient struct {
	client  ProductServiceClient
	timeout time.Duration
}

func newProductsGrpcClient(addr string, timeout time.Duration) (*ProductsGrpcClient, error) {
	conn, err := googlegrpc.NewClient(
		addr,
		googlegrpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &ProductsGrpcClient{client: NewProductServiceClient(conn), timeout: timeout}, nil
}

func (c *ProductsGrpcClient) GetProductByID(ctx context.Context, id string) (*GetProductByIDResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// forward the caller's request ID so products can correlate its logs with this request.
	if requestID, ok := logger.RequestIDFromContext(ctx); ok {
		ctx = metadata.AppendToOutgoingContext(ctx, logger.RequestIDMetadataKey, requestID)
	}

	return c.client.GetProductByID(ctx, &GetProductByIDRequest{Id: id})
}

func MustNewProductsGrpcClient(addr string, timeout time.Duration) *ProductsGrpcClient {
	client, err := newProductsGrpcClient(addr, timeout)
	if err != nil {
		log.Fatalf("failed to connect to products service at %s: %v", addr, err)
	}
	return client
}
