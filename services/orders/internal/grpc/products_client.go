package grpc

import (
	"context"
	"log"

	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ProductsGrpcClient adapts the generated ProductServiceClient to the
// service.ProductsClient interface, hiding gRPC boilerplate from the service layer.
type ProductsGrpcClient struct {
	client ProductServiceClient
}

func NewProductsGrpcClient(addr string) (*ProductsGrpcClient, error) {
	conn, err := googlegrpc.NewClient(
		addr,
		googlegrpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &ProductsGrpcClient{client: NewProductServiceClient(conn)}, nil
}

func (c *ProductsGrpcClient) GetProductByID(ctx context.Context, id string) (*GetProductByIDResponse, error) {
	return c.client.GetProductByID(ctx, &GetProductByIDRequest{Id: id})
}

func MustNewProductsGrpcClient(addr string) *ProductsGrpcClient {
	client, err := NewProductsGrpcClient(addr)
	if err != nil {
		log.Fatalf("failed to connect to products service at %s: %v", addr, err)
	}
	return client
}
