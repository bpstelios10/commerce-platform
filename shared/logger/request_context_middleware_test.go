package logger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// To be used as BeforeEach
func setupOrderHandlerWithContextTest(t *testing.T) *chi.Mux {
	t.Helper()

	base := New(Config{
		Service: "orders",
		Env:     "local",
		Level:   0, // InfoLevel
	})

	r := chi.NewRouter()
	r.Use(RequestContextMiddleware(base))
	r.Get("/dummy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	return r
}

func TestRequestContextMiddleware_WhenRequestIdProvided_ReturnsSameInResponseHeaders(t *testing.T) {
	r := setupOrderHandlerWithContextTest(t)

	req := httptest.NewRequest(http.MethodGet, "/dummy", nil)
	req.Header.Set("X-Request-Id", "test-request-id")
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNoContent, res.Code)
	assert.Equal(t, "test-request-id", res.Header().Get("X-Request-Id"))
}

func TestRequestContextMiddleware_WhenRequestIdNotProvided_ReturnsNewInResponseHeadersAndIsUuid(t *testing.T) {
	r := setupOrderHandlerWithContextTest(t)

	req := httptest.NewRequest(http.MethodGet, "/dummy", nil)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNoContent, res.Code)

	reqID := res.Header().Get("X-Request-Id")
	assert.NotEmpty(t, reqID)

	parsed, err := uuid.Parse(reqID)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, parsed)
}

func TestRequestContextMiddleware_AttachesRequestIDToContext(t *testing.T) {
	base := New(Config{Service: "orders", Env: "local", Level: 0})

	var requestIDFromCtx string
	r := chi.NewRouter()
	r.Use(RequestContextMiddleware(base))
	r.Get("/dummy", func(w http.ResponseWriter, r *http.Request) {
		requestIDFromCtx, _ = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/dummy", nil)
	req.Header.Set("X-Request-Id", "test-request-id")
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, "test-request-id", requestIDFromCtx)
}

func TestRequestContextMiddleware_LogsOneCanonicalLineWithMethodPathStatusAndDuration(t *testing.T) {
	out := captureStdout(t, func() {
		base := New(Config{Service: "orders", Env: "dev", Level: 0})

		r := chi.NewRouter()
		r.Use(RequestContextMiddleware(base))
		r.Get("/dummy/{id}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := httptest.NewRequest(http.MethodGet, "/dummy/123", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)
	})

	var entry map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &entry))
	assert.Equal(t, "request completed", entry["message"])
	assert.Equal(t, http.MethodGet, entry["method"])
	assert.Equal(t, "/dummy/123", entry["path"])
	assert.Equal(t, float64(http.StatusTeapot), entry["status"])
	assert.Contains(t, entry, "duration")
	assert.Contains(t, entry, "request_id")
}

func TestRequestContextMiddleware_WhenHandlerNeverCallsWriteHeader_LogsStatus200(t *testing.T) {
	out := captureStdout(t, func() {
		base := New(Config{Service: "orders", Env: "dev", Level: 0})

		r := chi.NewRouter()
		r.Use(RequestContextMiddleware(base))
		r.Get("/dummy", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})

		req := httptest.NewRequest(http.MethodGet, "/dummy", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)
	})

	var entry map[string]any
	assert.NoError(t, json.Unmarshal([]byte(out), &entry))
	assert.Equal(t, float64(http.StatusOK), entry["status"])
}
