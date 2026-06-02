package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestGetHealth_Returns200(t *testing.T) {
	handler := NewHealthHandler()

	r := chi.NewRouter()

	handler.RegisterRoutes(r)

	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "healthy", res.Body.String())
}
