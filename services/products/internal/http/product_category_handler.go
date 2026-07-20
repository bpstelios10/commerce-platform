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
	categories := h.productCategoryService.GetProductCategories()

	log().Info("product categories retrieved", "count", len(categories))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}
