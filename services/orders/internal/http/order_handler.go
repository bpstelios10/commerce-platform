package http

import (
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/services/orders/internal/service"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

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
	r.Post("/orders", h.CreateOrder)
	r.Put("/orders/{id}", h.UpdateOrder)
	r.Delete("/orders/{id}", h.DeleteProduct)
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

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.Warn("validation error occurred while creating order", "error", err)
		HandleError(w, service.ErrInvalidOrder)
		return
	}

	if err = validateCreateOrder(req); err != nil {
		HandleError(w, err)
		return
	}

	slog.Info("create order request received", "request", req)
	h.service.CreateOrder(req.ID, req.ProductID, req.Quantity)

	w.WriteHeader(http.StatusCreated)
}

func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	slog.Info("update order request received", "OrderId", id)

	var req UpdateOrderRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.Warn("validation error occurred while updating order", "error", err)
		HandleError(w, service.ErrInvalidOrder)
		return
	}

	// normalize status
	req.Status = order.OrderStatus(strings.ToUpper(string(req.Status)))
	if err = validateUpdateOrder(req); err != nil {
		HandleError(w, err)
		return
	}

	slog.Info("updated fields:", "request", req)
	h.service.UpdateOrder(id, req.ProductID, req.Quantity, req.Status)

	w.WriteHeader(http.StatusOK)
}

func (h *OrderHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	h.service.DeleteOrder(id)

	w.WriteHeader(http.StatusNoContent)
}
