package httpx

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.GetHealth)
}

func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("healthy"))
}
