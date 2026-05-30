package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

func (h *AdminHandler) RegisterRoutes(r chi.Router) {
	r.Get("/admin", h.GetAdmin)
}

func (h *AdminHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("admin"))
}
