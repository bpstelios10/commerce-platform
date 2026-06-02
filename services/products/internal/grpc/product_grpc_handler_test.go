package grpc

import (
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"
	context "context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestGetProductByID_WhenProductExists_ReturnsProduct(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := service.NewProductService(repo)
	handler := NewProductGrpcHandler(svc)

	server := grpc.NewServer()
	RegisterProductServiceServer(server, handler)
	lis := bufconn.Listen(1024 * 1024)
	go server.Serve(lis)

	ctx := context.Background()
	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	assert.NoError(t, err)

	client := NewProductServiceClient(conn)
	res, err := client.GetProductByID(
		context.Background(),
		&GetProductByIDRequest{
			Id: "1",
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, "1", res.Id)
}

func TestGetProductByID_WhenProductNotExists_ReturnsError(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := service.NewProductService(repo)
	handler := NewProductGrpcHandler(svc)

	server := grpc.NewServer()
	RegisterProductServiceServer(server, handler)
	lis := bufconn.Listen(1024 * 1024)
	go server.Serve(lis)

	ctx := context.Background()
	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	assert.NoError(t, err)

	client := NewProductServiceClient(conn)
	res, err := client.GetProductByID(
		context.Background(),
		&GetProductByIDRequest{
			Id: "999",
		},
	)

	assert.Error(t, err)
	assert.Nil(t, res)
}
