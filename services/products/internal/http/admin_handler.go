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
	ctx := r.Context()
	logger := log(ctx)

	var req CreateProductRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Warn().Err(err).Msg("validation error occurred while creating product")
		HandleError(ctx, w, service.ErrInvalidProduct)
		return
	}

	if err := validateCreateProduct(ctx, req); err != nil {
		HandleError(ctx, w, err)
		return
	}

	logger.Info().Interface("request", req).Msg("create product request received")

	p, err := h.adminService.CreateProduct(ctx, req.Name, req.Category, req.Price, *req.Stock)
	if err != nil {
		HandleError(ctx, w, err)
		return
	}

	HandlePostResponse(ctx, w, http.StatusCreated, p, "/products/"+p.ID.String())
}

func (h *AdminHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log(ctx)

	id := chi.URLParam(r, "id")
	validUUID, err := validation.GetValidUUID(id)
	if err != nil {
		HandleError(ctx, w, err)
		return
	}
	logger.Info().Str("product_id", validUUID.String()).Msg("update product request received")

	var req UpdateProductRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Warn().Err(err).Msg("validation error occurred while updating product")
		HandleError(ctx, w, service.ErrInvalidProduct)
		return
	}

	if err := validateUpdateProduct(ctx, req); err != nil {
		HandleError(ctx, w, err)
		return
	}

	logger.Info().Interface("request", req).Msg("update product")

	p, err := h.adminService.UpdateProduct(ctx, validUUID, req.Name, req.Category, req.Price, *req.Stock)
	if err != nil {
		HandleError(ctx, w, err)
		return
	}

	HandleResponseWithBody(ctx, w, http.StatusOK, p)
}

func (h *AdminHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log(ctx)

	id := chi.URLParam(r, "id")
	validUUID, err := validation.GetValidUUID(id)
	if err != nil {
		HandleError(ctx, w, err)
		return
	}
	logger.Info().Str("product_id", validUUID.String()).Msg("delete product request received")

	h.adminService.DeleteProduct(ctx, validUUID)

	HandleResponse(ctx, w, http.StatusNoContent)
}
