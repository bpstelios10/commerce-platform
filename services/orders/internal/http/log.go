package http

import (
	"commerce-platform/shared/logger"
	"log/slog"
)

func log() *slog.Logger {
	return logger.GetLogger("http")
}
