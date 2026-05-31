package handler

import (
	"commerce-platform/services/products/internal/service"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	service *service.ProductService
}

func NewProductHandler(service *service.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

func (h *ProductHandler) RegisterRoutes(r chi.Router) {
	r.Get("/products", h.GetProducts)
	r.Get("/products/{id}", h.GetProduct)
}

func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	products := h.service.GetProducts()

	slog.Info(
		"products retrieved",
		"count", len(products),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	idPathParam := chi.URLParam(r, "id")
	product, found := h.service.GetProductByID(idPathParam)

	if !found {
		slog.Warn(
			"product not found",
			"productId", idPathParam,
		)

		http.NotFound(w, r)
		return
	}

	slog.Info(
		"product lookup",
		"productId", idPathParam,
		"found", found,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}
