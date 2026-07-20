package httpx

import (
	"encoding/json"
	"net/http"

	"commerce-platform/services/products/internal/service"
	"commerce-platform/services/products/internal/validation"

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
		log().Warn("validation error occurred while creating product", "error", err)
		HandleError(w, service.ErrInvalidProduct)
		return
	}

	if err := validateCreateProduct(req); err != nil {
		HandleError(w, err)
		return
	}

	log().Info("create product request received", "request", req)

	p, err := h.adminService.CreateProduct(req.Name, req.Category, req.Price, *req.Stock)
	if err != nil {
		HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", "/products/"+p.ID.String())
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *AdminHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	validUUID, err := validation.GetValidUUID(id)
	if err != nil {
		HandleError(w, err)
		return
	}
	log().Info("update product request received", "ProductId", validUUID)

	var req UpdateProductRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log().Warn("validation error occurred while updating product", "error", err)
		HandleError(w, service.ErrInvalidProduct)
		return
	}

	if err := validateUpdateProduct(req); err != nil {
		HandleError(w, err)
		return
	}

	log().Info("update product", "request", req)

	p, err := h.adminService.UpdateProduct(validUUID, req.Name, req.Category, req.Price, *req.Stock)
	if err != nil {
		HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(p)
}

func (h *AdminHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	validUUID, err := validation.GetValidUUID(id)
	if err != nil {
		HandleError(w, err)
		return
	}
	log().Info("delete product request received", "ProductId", validUUID)

	h.adminService.DeleteProduct(validUUID)

	w.WriteHeader(http.StatusNoContent)
}
