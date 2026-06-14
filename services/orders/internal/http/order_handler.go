package http

import (
	"commerce-platform/services/orders/internal/service"
	"commerce-platform/services/orders/internal/validation"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) RegisterRoutes(r chi.Router) {
	r.Get("/orders", h.GetOrders)
	r.Get("/orders/{id}", h.GetOrder)
	r.Post("/orders", h.CreateOrder)
	r.Put("/orders/{id}", h.UpdateOrder)
	r.Delete("/orders/{id}", h.DeleteOrder)
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	orders := h.orderService.GetOrders()

	slog.Info("orders retrieved", "count", len(orders))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := validation.GetValidUUID(idParam)
	if err != nil {
		HandleError(w, err)
		return
	}

	o, err := h.orderService.GetOrderByID(id)

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
	if err = h.orderService.CreateOrder(r.Context(), uuid.MustParse(req.ID), req.ProductID, req.Quantity); err != nil {
		HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := validation.GetValidUUID(idParam)
	if err != nil {
		HandleError(w, err)
		return
	}
	slog.Info("update order request received", "orderId", id)

	var req UpdateOrderRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.Warn("validation error occurred while updating order", "error", err)
		HandleError(w, service.ErrInvalidOrder)
		return
	}

	// normalize status - to uppercase
	req.Status = req.Status.Normalize()
	if err = validateUpdateOrder(req); err != nil {
		HandleError(w, err)
		return
	}

	slog.Info("update order", "request", req)
	if err = h.orderService.UpdateOrder(r.Context(), id, req.ProductID, req.Quantity, req.Status); err != nil {
		HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := validation.GetValidUUID(idParam)
	if err != nil {
		HandleError(w, err)
		return
	}

	h.orderService.DeleteOrder(id)

	w.WriteHeader(http.StatusNoContent)
}
