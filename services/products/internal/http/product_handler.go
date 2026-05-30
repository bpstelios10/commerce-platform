package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ProductHandler struct{}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}

func (h *ProductHandler) RegisterRoutes(r chi.Router) {
	r.Get("/products/{id}", h.GetProduct)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("product"))
}
