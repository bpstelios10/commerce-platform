package logger

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestLevelFromEnv_WhenUnset_ReturnsFallback(t *testing.T) {
	level := LevelFromEnv("COMMERCE_TEST_LOG_LEVEL", zerolog.InfoLevel)

	assert.Equal(t, zerolog.InfoLevel, level)
}

func TestLevelFromEnv_WhenValid_ReturnsParsedLevel(t *testing.T) {
	t.Setenv("COMMERCE_TEST_LOG_LEVEL", "debug")

	level := LevelFromEnv("COMMERCE_TEST_LOG_LEVEL", zerolog.InfoLevel)

	assert.Equal(t, zerolog.DebugLevel, level)
}

func TestLevelFromEnv_WhenInvalid_ReturnsFallback(t *testing.T) {
	t.Setenv("COMMERCE_TEST_LOG_LEVEL", "not-a-level")

	level := LevelFromEnv("COMMERCE_TEST_LOG_LEVEL", zerolog.InfoLevel)

	assert.Equal(t, zerolog.InfoLevel, level)
}
