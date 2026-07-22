package httpx

import (
	"commerce-platform/shared/logger"
	"context"

	"github.com/rs/zerolog"
)

func log(ctx context.Context) zerolog.Logger {
	return logger.GetLogger(ctx, "http")
}
