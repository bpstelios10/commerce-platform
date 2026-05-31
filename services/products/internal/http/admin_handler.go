package handler

import (
	"commerce-platform/services/products/internal/service"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

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
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.validate(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.Info(
		"create product request received",
		"request", req,
	)

	h.adminService.CreateProduct(req.ID, req.Name, req.Price)

	w.WriteHeader(http.StatusCreated)
}

func (h *AdminHandler) validate(req CreateProductRequest) error {
	var errs []string

	if req.ID == "" {
		errs = append(errs, "id is required")
	}
	if req.Name == "" {
		errs = append(errs, "name is required")
	}
	if req.Price <= 0 {
		errs = append(errs, "price must be > 0")
	}

	if len(errs) > 0 {
		msg := strings.Join(errs, "; ")

		slog.Warn("invalid create product request", "error", msg)

		return fmt.Errorf("%s", msg)
	}

	return nil
}
