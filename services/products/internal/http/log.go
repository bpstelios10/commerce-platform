package httpx

import "log/slog"

func log() *slog.Logger {
	return slog.Default().With("component", "http")
}
