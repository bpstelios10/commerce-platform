package http

import (
	"commerce-platform/services/products/internal/service"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ProductCategoryHandler struct {
	productCategoryService *service.ProductCategoryService
}

func NewProductCategoryHandler(productCategoryService *service.ProductCategoryService) *ProductCategoryHandler {
	return &ProductCategoryHandler{
		productCategoryService: productCategoryService,
	}
}

func (h *ProductCategoryHandler) RegisterRoutes(r chi.Router) {
	r.Get("/products/categories", h.GetProductCategories)
}

func (h *ProductCategoryHandler) GetProductCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log(ctx)

	categories, err := h.productCategoryService.GetProductCategories(ctx)
	if err != nil {
		HandleError(ctx, w, fmt.Errorf("get_product_categories: %w", err))
		return
	}

	logger.Info().Int("count", len(categories)).Msg("product categories retrieved")

	HandleResponseWithBody(ctx, w, http.StatusOK, categories)
}
