package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Service string
	Env     string
	Level   slog.Level
}

func New(cfg Config) *slog.Logger {
	if cfg.Service == "" {
		cfg.Service = "N/D"
	}

	if cfg.Env == "" {
		cfg.Env = "local"
	}

	if cfg.Level == 0 {
		cfg.Level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: true,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				if src, ok := a.Value.Any().(*slog.Source); ok {
					return slog.String(slog.SourceKey, filepath.Base(src.File)+":"+strconv.Itoa(src.Line))
				}
			}
			return a
		},
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

func NewAndSetDefault(cfg Config) {
	logger := New(cfg)
	slog.SetDefault(logger)
}

func GetLogger(component string) *slog.Logger {
	if component == "" {
		component = "N/D"
	}

	return slog.Default().With("component", component)
}
