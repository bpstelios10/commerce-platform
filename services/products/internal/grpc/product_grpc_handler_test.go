package grpc

import (
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"
	"context"
	"net"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func setupProductHandlerTest(t *testing.T) ProductServiceClient {
	t.Helper()
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

	t.Cleanup(func() { server.Stop() })
	t.Cleanup(func() { conn.Close() })

	return client
}

func TestGetProductByID_WhenProductExists_ReturnsProduct(t *testing.T) {
	client := setupProductHandlerTest(t)

	res, err := client.GetProductByID(
		context.Background(),
		&GetProductByIDRequest{
			Id: repository.FirstUUID.String(),
		},
	)

	assert.NoError(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.OK, st.Code())
	assert.Equal(t, repository.FirstUUID.String(), res.Id)
	assert.Equal(t, "MacBook Pro", res.Name)
	assert.Equal(t, "ACCESSORY", res.Category)
	assert.Equal(t, 2500.0, res.Price)
}

func TestGetProductByID_WhenProductNotExists_ReturnsError(t *testing.T) {
	client := setupProductHandlerTest(t)
	id, _ := uuid.NewV7()

	res, err := client.GetProductByID(
		context.Background(),
		&GetProductByIDRequest{
			Id: id.String(),
		},
	)

	assert.Nil(t, res)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Equal(t, "product not found", st.Message())
}

func TestGetProductByID_WhenBadUUID_ReturnsError(t *testing.T) {
	client := setupProductHandlerTest(t)

	res, err := client.GetProductByID(
		context.Background(),
		&GetProductByIDRequest{
			Id: "1234",
		},
	)

	assert.Nil(t, res)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Equal(t, "invalid UUID", st.Message())
}
