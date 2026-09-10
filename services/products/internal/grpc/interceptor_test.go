package grpc

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"commerce-platform/shared/logger"

	"github.com/stretchr/testify/assert"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	assert.NoError(t, err)

	os.Stdout = w

	fn()

	assert.NoError(t, w.Close())
	os.Stdout = oldStdout

	bytes, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.NoError(t, r.Close())

	return string(bytes)
}

func TestLoggingUnaryInterceptor_WhenNoRequestIDInMetadata_GeneratesOneAndLogsCompletion(t *testing.T) {
	var requestIDInHandlerCtx string
	handler := func(ctx context.Context, req any) (any, error) {
		requestIDInHandlerCtx, _ = logger.RequestIDFromContext(ctx)
		return nil, nil
	}

	out := captureStdout(t, func() {
		baseLogger := logger.New(logger.Config{Service: "products", Env: "dev"})
		interceptor := LoggingUnaryInterceptor(baseLogger)

		_, err := interceptor(context.Background(), nil, &googlegrpc.UnaryServerInfo{FullMethod: "/product.ProductService/GetProductByID"}, handler)
		assert.NoError(t, err)
	})

	assert.NotEmpty(t, requestIDInHandlerCtx)

	var entry map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &entry))
	assert.Equal(t, requestIDInHandlerCtx, entry["request_id"])
	assert.Equal(t, "/product.ProductService/GetProductByID", entry["method"])
	assert.Equal(t, codes.OK.String(), entry["code"])
	assert.Contains(t, entry, "duration")
	assert.Equal(t, "rpc completed", entry["message"])
}

func TestLoggingUnaryInterceptor_WhenRequestIDInMetadata_ReusesIt(t *testing.T) {
	var requestIDInHandlerCtx string
	handler := func(ctx context.Context, req any) (any, error) {
		requestIDInHandlerCtx, _ = logger.RequestIDFromContext(ctx)
		return nil, nil
	}

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs(logger.RequestIDMetadataKey, "test-request-id"),
	)

	out := captureStdout(t, func() {
		baseLogger := logger.New(logger.Config{Service: "products", Env: "dev"})
		interceptor := LoggingUnaryInterceptor(baseLogger)

		_, err := interceptor(ctx, nil, &googlegrpc.UnaryServerInfo{}, handler)
		assert.NoError(t, err)
	})

	assert.Equal(t, "test-request-id", requestIDInHandlerCtx)

	var entry map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &entry))
	assert.Equal(t, "test-request-id", entry["request_id"])
}

func TestLoggingUnaryInterceptor_WhenHandlerReturnsError_LogsItsStatusCode(t *testing.T) {
	handler := func(ctx context.Context, req any) (any, error) {
		return nil, status.Error(codes.NotFound, "product not found")
	}

	out := captureStdout(t, func() {
		baseLogger := logger.New(logger.Config{Service: "products", Env: "dev"})
		interceptor := LoggingUnaryInterceptor(baseLogger)

		_, err := interceptor(context.Background(), nil, &googlegrpc.UnaryServerInfo{}, handler)
		assert.Error(t, err)
	})

	var entry map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &entry))
	assert.Equal(t, codes.NotFound.String(), entry["code"])
}
