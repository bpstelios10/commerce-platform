package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

type Config struct {
	Service string
	Env     string
	Level   slog.Level
}

func New(cfg Config) zerolog.Logger {
	if cfg.Service == "" {
		cfg.Service = "N/D"
	}

	if cfg.Env == "" {
		cfg.Env = "local"
	}

	if cfg.Level == 0 {
		cfg.Level = slog.LevelInfo
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

func GetLogger(component string) *slog.Logger {
	if component == "" {
		component = "N/D"
	}

	return slog.Default().With("component", component)
}
