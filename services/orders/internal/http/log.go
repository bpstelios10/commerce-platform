package http

import "log/slog"

func log() *slog.Logger {
	return slog.Default().With("component", "orders.http")
}
