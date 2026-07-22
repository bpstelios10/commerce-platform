package logger

import (
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
