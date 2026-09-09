package grpc

import (
	"context"
	"testing"

	"commerce-platform/shared/logger"

	"github.com/stretchr/testify/assert"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// capturingProductServiceClient records the outgoing metadata it was called with,
// instead of making a real network call.
type capturingProductServiceClient struct {
	capturedMetadata metadata.MD
}

func (c *capturingProductServiceClient) GetProductByID(ctx context.Context, in *GetProductByIDRequest, opts ...googlegrpc.CallOption) (*GetProductByIDResponse, error) {
	c.capturedMetadata, _ = metadata.FromOutgoingContext(ctx)
	return &GetProductByIDResponse{Id: in.Id}, nil
}

func TestGetProductByID_WhenRequestIDInContext_ForwardsItAsMetadata(t *testing.T) {
	client := &capturingProductServiceClient{}
	c := &ProductsGrpcClient{client: client}
	ctx := logger.ContextWithRequestID(context.Background(), "test-request-id")

	_, err := c.GetProductByID(ctx, "some-id")

	assert.NoError(t, err)
	assert.Equal(t, []string{"test-request-id"}, client.capturedMetadata.Get(logger.RequestIDMetadataKey))
}

func TestGetProductByID_WhenNoRequestIDInContext_DoesNotAddMetadata(t *testing.T) {
	client := &capturingProductServiceClient{}
	c := &ProductsGrpcClient{client: client}

	_, err := c.GetProductByID(context.Background(), "some-id")

	assert.NoError(t, err)
	assert.Empty(t, client.capturedMetadata.Get(logger.RequestIDMetadataKey))
}
