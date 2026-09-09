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
	"google.golang.org/grpc/metadata"
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

func TestLoggingUnaryInterceptor_WhenNoRequestIDInMetadata_GeneratesOne(t *testing.T) {
	var requestIDInHandlerCtx string
	handler := func(ctx context.Context, req any) (any, error) {
		requestIDInHandlerCtx, _ = logger.RequestIDFromContext(ctx)
		l := logger.GetLogger(ctx, "test")
		l.Info().Msg("handled")
		return nil, nil
	}

	out := captureStdout(t, func() {
		baseLogger := logger.New(logger.Config{Service: "products", Env: "dev"})
		interceptor := LoggingUnaryInterceptor(baseLogger)

		_, err := interceptor(context.Background(), nil, &googlegrpc.UnaryServerInfo{}, handler)
		assert.NoError(t, err)
	})

	assert.NotEmpty(t, requestIDInHandlerCtx)

	var entry map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &entry))
	assert.Equal(t, requestIDInHandlerCtx, entry["request_id"])
}

func TestLoggingUnaryInterceptor_WhenRequestIDInMetadata_ReusesIt(t *testing.T) {
	var requestIDInHandlerCtx string
	handler := func(ctx context.Context, req any) (any, error) {
		requestIDInHandlerCtx, _ = logger.RequestIDFromContext(ctx)
		l := logger.GetLogger(ctx, "test")
		l.Info().Msg("handled")
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
