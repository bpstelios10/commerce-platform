package grpc

import (
	"context"
	"time"

	"commerce-platform/shared/logger"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LoggingUnaryInterceptor attaches a request-scoped logger to the context of every
// unary RPC, mirroring shared/logger.RequestContextMiddleware for HTTP: it reuses the
// caller's request ID when propagated via metadata, or generates one otherwise. It also
// emits one canonical "rpc completed" log line per call, independent of whatever the
// handler itself logs.
func LoggingUnaryInterceptor(baseLogger zerolog.Logger) googlegrpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *googlegrpc.UnaryServerInfo, handler googlegrpc.UnaryHandler) (any, error) {
		start := time.Now()

		requestID := requestIDFromIncomingMetadata(ctx)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		reqLogger := baseLogger.With().Str("request_id", requestID).Logger()
		ctx = reqLogger.WithContext(ctx)
		ctx = logger.ContextWithRequestID(ctx, requestID)

		resp, err := handler(ctx, req)

		reqLogger.Info().
			Str("method", info.FullMethod).
			Str("code", status.Code(err).String()).
			Dur("duration", time.Since(start)).
			Msg("rpc completed")

		return resp, err
	}
}

func requestIDFromIncomingMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get(logger.RequestIDMetadataKey)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}
