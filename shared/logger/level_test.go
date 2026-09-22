package logger

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestLevelFromEnvWithFallback_WhenUnset_ReturnsFallback(t *testing.T) {
	level := LevelFromEnvWithFallback("COMMERCE_TEST_LOG_LEVEL", zerolog.InfoLevel)

	assert.Equal(t, zerolog.InfoLevel, level)
}

func TestLevelFromEnvWithFallback_WhenValid_ReturnsParsedLevel(t *testing.T) {
	t.Setenv("COMMERCE_TEST_LOG_LEVEL", "debug")

	level := LevelFromEnvWithFallback("COMMERCE_TEST_LOG_LEVEL", zerolog.InfoLevel)

	assert.Equal(t, zerolog.DebugLevel, level)
}

func TestLevelFromEnvWithFallback_WhenInvalid_ReturnsFallback(t *testing.T) {
	t.Setenv("COMMERCE_TEST_LOG_LEVEL", "not-a-level")

	level := LevelFromEnvWithFallback("COMMERCE_TEST_LOG_LEVEL", zerolog.InfoLevel)

	assert.Equal(t, zerolog.InfoLevel, level)
}

func TestLevelFromEnv_WhenValid_ReturnsParsedLevel(t *testing.T) {
	t.Setenv("COMMERCE_TEST_LOG_LEVEL", "debug")

	level := LevelFromEnv("COMMERCE_TEST_LOG_LEVEL")

	assert.Equal(t, zerolog.DebugLevel, level)
}

func TestLevelFromEnv_WhenInvalid_ReturnsNoLevel(t *testing.T) {
	t.Setenv("COMMERCE_TEST_LOG_LEVEL", "not-a-level")

	level := LevelFromEnv("COMMERCE_TEST_LOG_LEVEL")

	assert.Equal(t, zerolog.NoLevel, level)
}
