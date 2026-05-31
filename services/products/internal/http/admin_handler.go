package handler

import (
	"commerce-platform/services/products/internal/service"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

func (h *AdminHandler) RegisterRoutes(r chi.Router) {
	r.Get("/admin", h.GetAdmin)
	r.Post("/admin/products", h.CreateProduct)
}

func (h *AdminHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("admin"))
}

func (h *AdminHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.Warn(
			"invalid create product request",
			"error", err,
		)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	slog.Info(
		"create product request received",
		"request", req,
	)

	h.adminService.CreateProduct(req.ID, req.Name, req.Price)

	w.WriteHeader(http.StatusCreated)
}
