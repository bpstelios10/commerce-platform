package grpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"commerce-platform/shared/logger"

	googlegrpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
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

func (c *ProductsGrpcClient) getProductByID(ctx context.Context, id string) (*GetProductByIDResponse, error) {
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
		panic(fmt.Sprintf("create products grpc client for %s: %v", addr, err))
	}
	return client
}

func (c *ProductsGrpcClient) ProductExists(ctx context.Context, productID string) (bool, error) {
	_, err := c.getProductByID(ctx, productID)
	if err != nil {
		wrapped := fmt.Errorf("get product %s from products service: %w", productID, err)
		switch status.Code(err) {
		case codes.NotFound:
			return false, nil
		case codes.DeadlineExceeded:
			return false, errors.Join(wrapped, context.DeadlineExceeded)
		case codes.Canceled:
			return false, errors.Join(wrapped, context.Canceled)
		}
		return false, wrapped
	}
	return true, nil
}
