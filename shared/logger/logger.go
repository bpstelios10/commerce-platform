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

func GetLogger(ctx context.Context, component string) zerolog.Logger {
	if component == "" {
		component = "N/D"
	}

	return zerolog.Ctx(ctx).With().Str("component", component).Logger()
}
