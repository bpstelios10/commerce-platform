package logger

import (
	"bufio"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

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

func TestNew_UsesFallbackFieldsAndInfoLevelByDefault(t *testing.T) {
	out := captureStdout(t, func() {
		logger := New(Config{})
		logger.Debug("debug-should-not-appear")
		logger.Info("hello-default")
	})

	assert.Contains(t, out, "level=INFO")
	assert.Contains(t, out, "msg=hello-default")
	assert.Contains(t, out, "service=unknown-service")
	assert.Contains(t, out, "env=local")
	assert.NotContains(t, out, "debug-should-not-appear")
}

func TestNew_UsesProvidedFieldsAndLevel(t *testing.T) {
	out := captureStdout(t, func() {
		logger := New(Config{
			Service: "orders",
			Env:     "dev",
			Level:   slog.LevelDebug,
		})
		logger.Debug("hello-debug")
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

	assert.Equal(t, "DEBUG", entry["level"])
	assert.Equal(t, "hello-debug", entry["msg"])
	assert.Equal(t, "orders", entry["service"])
	assert.Equal(t, "dev", entry["env"])
}
