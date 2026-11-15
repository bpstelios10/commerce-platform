package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"commerce-platform/shared/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	googlegrpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
)

// capturingProductServiceClient records the outgoing metadata it was called with,
// instead of making a real network call.
type capturingProductServiceClient struct {
	capturedMetadata metadata.MD
	deadline         time.Time
	hasDeadline      bool
	err              error
}

func (c *capturingProductServiceClient) GetProductByID(
	ctx context.Context,
	in *GetProductByIDRequest,
	_ ...googlegrpc.CallOption,
) (*GetProductByIDResponse, error) {
	c.capturedMetadata, _ = metadata.FromOutgoingContext(ctx)
	c.deadline, c.hasDeadline = ctx.Deadline()
	return &GetProductByIDResponse{Id: in.Id}, c.err
}

func TestGetProductByID_WhenRequestIDInContext_ForwardsItAsMetadata(t *testing.T) {
	fake := &capturingProductServiceClient{}
	c := &ProductsGrpcClient{client: fake, timeout: 500 * time.Millisecond}
	ctx := logger.ContextWithRequestID(context.Background(), "test-request-id")

	_, err := c.getProductByID(ctx, "some-id")

	assert.NoError(t, err)
	assert.Equal(t, []string{"test-request-id"}, fake.capturedMetadata.Get(logger.RequestIDMetadataKey))
}

func TestGetProductByID_WhenNoRequestIDInContext_DoesNotAddMetadata(t *testing.T) {
	fake := &capturingProductServiceClient{}
	c := &ProductsGrpcClient{client: fake}

	_, err := c.getProductByID(context.Background(), "some-id")

	assert.NoError(t, err)
	assert.Empty(t, fake.capturedMetadata.Get(logger.RequestIDMetadataKey))
}

func TestGetProductByID_WhenCallerDeadlineShorter_KeepsIt(t *testing.T) {
	fake := &capturingProductServiceClient{}
	c := &ProductsGrpcClient{client: fake, timeout: 2 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	want, _ := ctx.Deadline()

	_, err := c.getProductByID(ctx, "some-id")

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

	resp, err := c.getProductByID(context.Background(), "some-id")

	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Nil(t, resp)
	assert.Less(t, time.Since(start), time.Second)
}

// Test Constructor
type slowProductServer struct {
	UnimplementedProductServiceServer
	delay time.Duration
}

func (s slowProductServer) GetProductByID(ctx context.Context, in *GetProductByIDRequest) (*GetProductByIDResponse, error) {
	select {
	case <-time.After(s.delay):
		return &GetProductByIDResponse{Id: in.Id}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestMustNewProductsGrpcClient_WhenInvalidAddress_Panics(t *testing.T) {
	assert.Panics(t, func() {
		MustNewProductsGrpcClient("%", time.Second)
	})
}

func TestProductExists_WhenRequestIDInContext_ReturnsTrue(t *testing.T) {
	fake := &capturingProductServiceClient{}
	c := &ProductsGrpcClient{client: fake, timeout: 500 * time.Millisecond}
	ctx := logger.ContextWithRequestID(context.Background(), "test-request-id")

	exists, err := c.ProductExists(ctx, "some-id")

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, []string{"test-request-id"}, fake.capturedMetadata.Get(logger.RequestIDMetadataKey))
}

func TestProductExists_WhenNotFound_ReturnsFalse(t *testing.T) {
	fake := &capturingProductServiceClient{err: status.Error(codes.NotFound, "product not found")}
	c := &ProductsGrpcClient{client: fake, timeout: 500 * time.Millisecond}
	ctx := logger.ContextWithRequestID(context.Background(), "test-request-id")

	exists, err := c.ProductExists(ctx, "some-id")

	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestProductExists_WhenError_ReturnsError(t *testing.T) {
	fake := &capturingProductServiceClient{err: errors.New("some error")}
	c := &ProductsGrpcClient{client: fake, timeout: 500 * time.Millisecond}
	ctx := logger.ContextWithRequestID(context.Background(), "test-request-id")

	exists, err := c.ProductExists(ctx, "some-id")

	assert.Error(t, err)
	assert.False(t, exists)
	assert.EqualError(t, err, "get product some-id from products service: some error")
}

func TestProductExists_WhenDeadlineExceeded_ReturnsError(t *testing.T) {
	fake := &capturingProductServiceClient{err: status.Error(codes.DeadlineExceeded, "deadline exceeded")}
	c := &ProductsGrpcClient{client: fake, timeout: 500 * time.Millisecond}
	ctx := logger.ContextWithRequestID(context.Background(), "test-request-id")

	exists, err := c.ProductExists(ctx, "some-id")

	assert.Error(t, err)
	assert.False(t, exists)
	assert.EqualError(t, err, "get product some-id from products service: rpc error: code = DeadlineExceeded desc = deadline exceeded\ncontext deadline exceeded")
}

func TestProductExists_WhenCanceled_ReturnsError(t *testing.T) {
	fake := &capturingProductServiceClient{err: status.Error(codes.Canceled, "canceled")}
	c := &ProductsGrpcClient{client: fake, timeout: 500 * time.Millisecond}
	ctx := logger.ContextWithRequestID(context.Background(), "test-request-id")

	exists, err := c.ProductExists(ctx, "some-id")

	assert.Error(t, err)
	assert.False(t, exists)
	assert.EqualError(t, err, "get product some-id from products service: rpc error: code = Canceled desc = canceled\ncontext canceled")
}
