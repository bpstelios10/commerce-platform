package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	if !isBodyValidAndDecoded(ctx, r.Body, w, &req, "create_product") {
		return
	}

	if err := validateCreateProduct(ctx, req); err != nil {
		HandleError(ctx, w, err)
		return
	}

	logger.Info().Interface("request", req).Msg("create product request received")

	p, err := h.adminService.CreateProduct(ctx, req.Name, req.Category, req.Description, req.Price)
	if err != nil {
		HandleError(ctx, w, fmt.Errorf("create_product: %w", err))
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
	if !isBodyValidAndDecoded(ctx, r.Body, w, &req, "update_product") {
		return
	}

	if err := validateUpdateProduct(ctx, req); err != nil {
		HandleError(ctx, w, err)
		return
	}

	logger.Info().Interface("request", req).Msg("update product")

	p, err := h.adminService.UpdateProduct(ctx, validUUID, req.Name, req.Category, req.Description, req.Price)
	if err != nil {
		HandleError(ctx, w, fmt.Errorf("update_product: %w", err))
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

	err = h.adminService.DeleteProduct(ctx, validUUID)
	if err != nil {
		HandleError(ctx, w, fmt.Errorf("delete_order: %w", err))
		return
	}

	HandleResponse(ctx, w, http.StatusNoContent)
}

func isBodyValidAndDecoded[T any](ctx context.Context, body io.ReadCloser, w http.ResponseWriter, req *T, action string) bool {
	body = http.MaxBytesReader(w, body, 1<<10)
	dec := json.NewDecoder(body)

	if err := dec.Decode(&req); err != nil {
		var e *http.MaxBytesError
		if errors.As(err, &e) {
			err = errors.Join(
				fmt.Errorf("decode %s request: %w", action, err),
				TooLongBodyErr{limit: 1 << 10},
			)
		} else {
			err = errors.Join(
				fmt.Errorf("decode %s request: %w", action, err),
				service.ErrInvalidProduct,
			)
		}
		HandleError(ctx, w, err)
		return false
	}

	// trailing JSON/content
	if dec.Decode(&struct{}{}) != io.EOF {
		err := fmt.Errorf("decode %s request: %w", action, service.ErrInvalidProduct)
		HandleError(ctx, w, err)
		return false
	}

	return true
}
