package grpc

import (
	"context"
	"testing"
	"time"

	"net"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	googlegrpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

func startProductServer(t *testing.T, srv ProductServiceServer) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	s := googlegrpc.NewServer()
	RegisterProductServiceServer(s, srv)
	go s.Serve(lis)
	t.Cleanup(s.Stop)
	return lis.Addr().String()
}

func TestMustNewProductsGrpcClient_CallsServer(t *testing.T) {
	c := MustNewProductsGrpcClient(startProductServer(t, slowProductServer{}), time.Second)

	resp, err := c.getProductByID(context.Background(), "some-id")

	require.NoError(t, err)
	assert.Equal(t, "some-id", resp.Id)
}

func TestGetProductByID_WhenServerTooSlow_ReturnsDeadlineExceeded(t *testing.T) {
	c, err := newProductsGrpcClient(startProductServer(t, slowProductServer{delay: time.Second}), 50*time.Millisecond)
	require.NoError(t, err)

	_, err = c.getProductByID(context.Background(), "some-id")

	assert.Equal(t, codes.DeadlineExceeded, status.Code(err))
}
