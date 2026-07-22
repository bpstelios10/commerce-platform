package httpx

import (
	"commerce-platform/services/products/internal/service"
	"encoding/json"
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

	categories := h.productCategoryService.GetProductCategories()

	logger.Info().Int("count", len(categories)).Msg("product categories retrieved")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}
