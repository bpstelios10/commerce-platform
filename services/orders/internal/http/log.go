package http

import (
	"context"

	"github.com/rs/zerolog"
)

func log(ctx context.Context) zerolog.Logger {
	return zerolog.Ctx(ctx).With().Str("component", "http").Logger()
}
