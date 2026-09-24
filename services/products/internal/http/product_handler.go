package http

import (
	"commerce-platform/services/products/internal/service"
	"commerce-platform/services/products/internal/validation"
	"fmt"
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
	ctx := r.Context()
	logger := log(ctx)

	products, err := h.productService.GetProducts(ctx)
	if err != nil {
		HandleError(ctx, w, fmt.Errorf("get_products: %w", err))
		return
	}

	logger.Info().Int("count", len(products)).Msg("products retrieved")

	HandleResponseWithBody(ctx, w, http.StatusOK, products)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log(ctx)

	idPathParam := chi.URLParam(r, "id")
	validUUID, err := validation.GetValidUUID(idPathParam)
	if err != nil {
		HandleError(ctx, w, err)
		return
	}

	product, err := h.productService.GetProductByID(ctx, validUUID)
	if err != nil {
		HandleError(ctx, w, fmt.Errorf("get_product: %w", err))
		return
	}

	logger.Info().Str("product_id", validUUID.String()).Msg("product was found")

	HandleResponseWithBody(ctx, w, http.StatusOK, product)
}

// TODO if i have a category but the other 2 filters dont much, then return something? maybe like 10 products
// and then the AI can ask for more params to re-run the search
func (h *ProductHandler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log(ctx)

	query := r.URL.Query().Get("query")
	category := r.URL.Query().Get("category")

	var maxPrice *float64
	if maxPriceParam := r.URL.Query().Get("maxPrice"); maxPriceParam != "" {
		parsed, err := strconv.ParseFloat(maxPriceParam, 64)
		if err != nil {
			logger.Debug().
				Str("operation", "search_products").
				Str("parameter", "maxPrice").
				Err(err).
				Msg("validation error")
			HandleError(ctx, w, ValidationError{Errors: []string{"maxPrice must be a valid number."}})
			return
		}
		maxPrice = &parsed
	}

	logger.Info().
		Str("query", query).
		Interface("max_price", maxPrice).
		Str("category", category).
		Msg("products search request")

	products, err := h.productService.SearchProducts(ctx, query, maxPrice, category)
	if err != nil {
		HandleError(ctx, w, fmt.Errorf("search_products: %w", err))
		return
	}

	logger.Info().Int("count", len(products)).Msg("products found")

	HandleResponseWithBody(ctx, w, http.StatusOK, products)
}
