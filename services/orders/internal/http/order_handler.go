package http

import (
	"commerce-platform/services/orders/internal/service"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) RegisterRoutes(r chi.Router) {
	r.Get("/orders", h.GetOrders)
	r.Get("/orders/{id}", h.GetOrder)
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	orders := h.service.GetOrders()

	slog.Info("orders retrieved", "count", len(orders))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	o, err := h.service.GetOrderByID(id)

	if err != nil {
		HandleError(w, err)
		return
	}

	slog.Info("order was found, with", "orderId", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)
}
