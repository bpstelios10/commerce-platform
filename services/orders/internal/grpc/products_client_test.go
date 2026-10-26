package grpc

import (
	"context"
	"testing"
	"time"

	"commerce-platform/shared/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// capturingProductServiceClient records the outgoing metadata it was called with,
// instead of making a real network call.
type capturingProductServiceClient struct {
	capturedMetadata metadata.MD
	deadline         time.Time
	hasDeadline      bool
}

func (c *capturingProductServiceClient) GetProductByID(
	ctx context.Context,
	in *GetProductByIDRequest,
	_ ...googlegrpc.CallOption,
) (*GetProductByIDResponse, error) {
	c.capturedMetadata, _ = metadata.FromOutgoingContext(ctx)
	c.deadline, c.hasDeadline = ctx.Deadline()
	return &GetProductByIDResponse{Id: in.Id}, nil
}

func TestGetProductByID_WhenRequestIDInContext_ForwardsItAsMetadata(t *testing.T) {
	fake := &capturingProductServiceClient{}
	c := &ProductsGrpcClient{client: fake, timeout: 500 * time.Millisecond}
	ctx := logger.ContextWithRequestID(context.Background(), "test-request-id")

	_, err := c.GetProductByID(ctx, "some-id")

	assert.NoError(t, err)
	assert.Equal(t, []string{"test-request-id"}, fake.capturedMetadata.Get(logger.RequestIDMetadataKey))
}

func TestGetProductByID_WhenNoRequestIDInContext_DoesNotAddMetadata(t *testing.T) {
	fake := &capturingProductServiceClient{}
	c := &ProductsGrpcClient{client: fake}

	_, err := c.GetProductByID(context.Background(), "some-id")

	assert.NoError(t, err)
	assert.Empty(t, fake.capturedMetadata.Get(logger.RequestIDMetadataKey))
}

func TestGetProductByID_WhenCallerDeadlineShorter_KeepsIt(t *testing.T) {
	fake := &capturingProductServiceClient{}
	c := &ProductsGrpcClient{client: fake, timeout: 2 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	want, _ := ctx.Deadline()

	_, err := c.GetProductByID(ctx, "some-id")

	require.NoError(t, err)
	assert.Equal(t, want, fake.deadline)
}

// TIMEOUT test
type blockingProductServiceClient struct{}

func (blockingProductServiceClient) GetProductByID(ctx context.Context, _ *GetProductByIDRequest, _ ...googlegrpc.CallOption) (*GetProductByIDResponse, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestGetProductByID_WhenTimeoutExceeded_ReturnsDeadlineExceeded(t *testing.T) {
	c := &ProductsGrpcClient{client: blockingProductServiceClient{}, timeout: 50 * time.Millisecond}
	start := time.Now()

	resp, err := c.GetProductByID(context.Background(), "some-id")

	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Nil(t, resp)
	assert.Less(t, time.Since(start), time.Second)
}
