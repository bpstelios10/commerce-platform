package httpx

import (
	"commerce-platform/services/products/internal/service"
	"commerce-platform/services/products/internal/validation"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (h *ProductHandler) RegisterRoutes(r chi.Router) {
	r.Get("/products", h.GetProducts)
	r.Get("/products/{id}", h.GetProduct)
}

func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	products := h.productService.GetProducts()

	slog.Info("products retrieved", "count", len(products))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	idPathParam := chi.URLParam(r, "id")
	validUUID, err := validation.GetValidUUID(idPathParam)
	if err != nil {
		HandleError(w, err)
		return
	}

	product, err := h.productService.GetProductByID(validUUID)

	if err != nil {
		HandleError(w, err)
		return
	}

	slog.Info("product was found, with", "productId", idPathParam)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}
