package logger

import (
	"log/slog"
	"os"
)

type Config struct {
	Service string
	Env     string
	Level   slog.Level
}

func New(cfg Config) *slog.Logger {
	if cfg.Service == "" {
		cfg.Service = "unknown-service"
	}

	if cfg.Env == "" {
		cfg.Env = "local"
	}

	if cfg.Level == 0 {
		cfg.Level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: cfg.Level,
	}

	var handler slog.Handler

	if cfg.Env == "local" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler).With(
		"service", cfg.Service,
		"env", cfg.Env,
	)

	return logger
}
