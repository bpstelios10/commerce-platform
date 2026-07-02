package httpx

import (
	"commerce-platform/services/products/internal/service"
	"commerce-platform/services/products/internal/validation"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

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
	r.Get("/products/search", h.SearchProducts)
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

func (h *ProductHandler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	category := r.URL.Query().Get("category")

	var maxPrice *float64
	if maxPriceParam := r.URL.Query().Get("maxPrice"); maxPriceParam != "" {
		parsed, err := strconv.ParseFloat(maxPriceParam, 64)
		if err != nil {
			HandleError(w, ValidationError{Errors: []string{"maxPrice must be a valid number."}})
			return
		}
		maxPrice = &parsed
	}

	slog.Info("products search request", "query", query, "maxPrice", maxPrice, "category", category)

	products := h.productService.SearchProducts(query, maxPrice, category)
	slog.Info("products found", "products", products)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}
