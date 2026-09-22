package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// LevelFromEnv reads envVar and parses it as a zerolog level name (e.g. "debug", "info",
// "warn", "error"). Falls back to fallback if the variable is unset or not a valid level.
func LevelFromEnvWithFallback(envVar string, fallback zerolog.Level) zerolog.Level {
	value := os.Getenv(envVar)
	if value == "" {
		return fallback
	}

	level, err := zerolog.ParseLevel(value)
	if err != nil {
		return fallback
	}

	return level
}

func LevelFromEnv(envVar string) zerolog.Level {
	return LevelFromEnvWithFallback(envVar, zerolog.NoLevel)
}
