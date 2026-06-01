package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"commerce-platform/services/products/internal/service"

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
	r.Put("/admin/products/{id}", h.UpdateProduct)
	r.Delete("/admin/products/{id}", h.DeleteProduct)
}

func (h *AdminHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("admin"))
}

func (h *AdminHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		HandleError(w, err)
		return
	}

	if err := validateCreateProduct(req); err != nil {
		HandleError(w, err)
		return
	}

	slog.Info("create product request received", "request", req)

	h.adminService.CreateProduct(req.ID, req.Name, req.Price)

	w.WriteHeader(http.StatusCreated)
}

func (h *AdminHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	slog.Info("update product request received", "requestId", id)

	var req UpdateProductRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		HandleError(w, err)
		return
	}

	if err := validateUpdateProduct(req); err != nil {
		HandleError(w, err)
		return
	}

	slog.Info("updated fields:", "request", req)

	h.adminService.UpdateProduct(id, req.Name, req.Price)

	w.WriteHeader(http.StatusOK)
}

func (h *AdminHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	h.adminService.DeleteProduct(id)

	w.WriteHeader(http.StatusNoContent)
}
