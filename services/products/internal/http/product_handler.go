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
	logger         *slog.Logger
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{
		logger:         log(),
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

	h.log().Info("products retrieved", "count", len(products))

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

	h.log().Info("product was found, with", "productId", idPathParam)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// TODO if i have a category but the other 2 filters dont much, then return something? maybe like 10 products
// and then the AI can ask for more params to re-run the search
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

	h.log().Info("products search request", "query", query, "maxPrice", maxPrice, "category", category)

	products := h.productService.SearchProducts(query, maxPrice, category)
	h.log().Info("products found", "products", products)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) log() *slog.Logger {
	// here i could add more things, related to this class only. or else just use log()
	return h.logger
}
