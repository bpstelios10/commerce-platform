package logger

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	assert.NoError(t, err)

	os.Stdout = w

	fn()

	assert.NoError(t, w.Close())
	os.Stdout = oldStdout

	bytes, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.NoError(t, r.Close())

	return string(bytes)
}

func TestNew_UsesFallbackFieldsAndDebugLevelByDefault(t *testing.T) {
	out := captureStdout(t, func() {
		logger := New(Config{})
		logger.Debug().Msg("debug-should-also-appear")
		logger.Info().Msg("hello-default")
	})

	assert.Contains(t, out, "hello-default")
	assert.Contains(t, out, "service=\x1b[0mN/D")
	assert.Contains(t, out, "env=\x1b[0mlocal")
	assert.NotContains(t, out, "component")
	assert.Contains(t, out, "debug-should-also-appear")
}

func TestNew_UsesProvidedFieldsAndLevel(t *testing.T) {
	out := captureStdout(t, func() {
		logger := New(Config{
			Service: "orders",
			Env:     "dev",
			Level:   zerolog.InfoLevel,
		})
		logger.Debug().Msg("debug-should-not-appear")
		logger.Info().Msg("hello-debug")
	})

	assert.NotContains(t, out, "debug-should-not-appear")

	scanner := bufio.NewScanner(strings.NewReader(out))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	assert.NoError(t, scanner.Err())
	assert.Len(t, lines, 1)

	var entry map[string]any
	err := json.Unmarshal([]byte(lines[0]), &entry)
	assert.NoError(t, err)

	assert.Equal(t, "info", entry["level"])
	assert.Equal(t, "hello-debug", entry["message"])
	assert.Equal(t, "orders", entry["service"])
	assert.Equal(t, "dev", entry["env"])
	assert.Contains(t, entry, "time")
	assert.Contains(t, entry, "caller")
}

func TestNewAndSetDefault_UsesProvidedFieldsAndLevel(t *testing.T) {
	out := captureStdout(t, func() {
		logger := New(Config{
			Service: "orders",
			Env:     "dev",
			Level:   zerolog.DebugLevel,
		})
		logger.Debug().Msg("hello-debug")
	})

	scanner := bufio.NewScanner(strings.NewReader(out))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	assert.NoError(t, scanner.Err())
	assert.Len(t, lines, 1)

	var entry map[string]any
	err := json.Unmarshal([]byte(lines[0]), &entry)
	assert.NoError(t, err)

	assert.Equal(t, "debug", entry["level"])
	assert.Equal(t, "hello-debug", entry["message"])
	assert.Equal(t, "orders", entry["service"])
	assert.Equal(t, "dev", entry["env"])
}

func TestGetLogger_whenLoggerWithComponent(t *testing.T) {
	out := captureStdout(t, func() {
		logger := New(Config{
			Service: "orders",
			Env:     "dev",
			Level:   zerolog.DebugLevel,
		})
		reqLogger := logger.With().Str("request_id", "X").Logger()
		ctx := reqLogger.WithContext(context.Background())

		l := GetLogger(ctx, "test")
		l.Debug().Msg("hello-test-debug")
	})

	scanner := bufio.NewScanner(strings.NewReader(out))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	assert.NoError(t, scanner.Err())
	assert.Len(t, lines, 1)

	var entry map[string]any
	err := json.Unmarshal([]byte(lines[0]), &entry)
	assert.NoError(t, err)

	assert.Equal(t, "debug", entry["level"])
	assert.Equal(t, "hello-test-debug", entry["message"])
	assert.Equal(t, "orders", entry["service"])
	assert.Equal(t, "test", entry["component"])
	assert.Equal(t, "dev", entry["env"])
}

func TestGetLogger_whenLoggerWithoutComponent(t *testing.T) {
	out := captureStdout(t, func() {
		logger := New(Config{
			Service: "orders",
			Env:     "dev",
			Level:   zerolog.DebugLevel,
		})
		reqLogger := logger.With().Str("request_id", "X").Logger()
		ctx := reqLogger.WithContext(context.Background())

		l := GetLogger(ctx, "")
		l.Debug().Msg("hello-test-debug")
	})

	scanner := bufio.NewScanner(strings.NewReader(out))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	assert.NoError(t, scanner.Err())
	assert.Len(t, lines, 1)

	var entry map[string]any
	err := json.Unmarshal([]byte(lines[0]), &entry)
	assert.NoError(t, err)

	assert.Equal(t, "debug", entry["level"])
	assert.Equal(t, "hello-test-debug", entry["message"])
	assert.Equal(t, "orders", entry["service"])
	assert.Equal(t, "N/D", entry["component"])
	assert.Equal(t, "dev", entry["env"])
}
