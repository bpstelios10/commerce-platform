package logger

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

type Config struct {
	Service string
	Env     string
	Level   zerolog.Level
}

func New(cfg Config) zerolog.Logger {
	if cfg.Service == "" {
		cfg.Service = "N/D"
	}

	if cfg.Env == "" {
		cfg.Env = "local"
	}

	var output io.Writer
	if cfg.Env == "local" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	} else {
		output = os.Stdout
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	logger := zerolog.New(output).With().
		Timestamp().
		Caller().
		Str("service", cfg.Service).
		Str("env", cfg.Env).
		Logger()
	zerolog.SetGlobalLevel(zerolog.Level(cfg.Level))

	return logger
}

// SetAsDefault installs logger as the process-wide fallback used by GetLogger(ctx, ...)
// when ctx has no request-scoped logger attached (e.g. gRPC calls without a logging
// interceptor, or startup code). Call this once from main, after New; New itself stays
// side-effect-free so it's safe to call repeatedly in tests.
func SetAsDefault(logger zerolog.Logger) {
	zerolog.DefaultContextLogger = &logger
}

func GetLogger(ctx context.Context, component string) zerolog.Logger {
	if component == "" {
		component = "N/D"
	}

	return zerolog.Ctx(ctx).With().Str("component", component).Logger()
}
