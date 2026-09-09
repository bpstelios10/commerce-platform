package logger

import "context"

// RequestIDMetadataKey is the gRPC metadata key used to propagate the request/
// correlation ID across service boundaries (gRPC lower-cases metadata keys).
const RequestIDMetadataKey = "x-request-id"

type ctxKey int

const requestIDKey ctxKey = iota

// ContextWithRequestID attaches requestID to ctx so it can be read back later,
// e.g. to forward it as an outgoing gRPC metadata value on the next hop.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext returns the request ID attached via ContextWithRequestID, if any.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDKey).(string)
	return requestID, ok
}
