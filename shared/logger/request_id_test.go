package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestIDFromContext_WhenNotSet_ReturnsFalse(t *testing.T) {
	requestID, ok := RequestIDFromContext(context.Background())

	assert.False(t, ok)
	assert.Empty(t, requestID)
}

func TestContextWithRequestID_ThenRequestIDFromContext_ReturnsSameValue(t *testing.T) {
	ctx := ContextWithRequestID(context.Background(), "test-request-id")

	requestID, ok := RequestIDFromContext(ctx)

	assert.True(t, ok)
	assert.Equal(t, "test-request-id", requestID)
}
