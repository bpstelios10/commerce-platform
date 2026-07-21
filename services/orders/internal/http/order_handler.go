package http

import (
	"commerce-platform/services/orders/internal/service"
	"commerce-platform/services/orders/internal/validation"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) RegisterRoutes(r chi.Router) {
	r.Get("/orders", h.GetOrders)
	r.Get("/orders/{id}", h.GetOrder)
	r.Post("/orders", h.CreateOrder)
	r.Put("/orders/{id}", h.UpdateOrder)
	r.Delete("/orders/{id}", h.DeleteOrder)
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log(ctx)

	orders := h.orderService.GetOrders()

	logger.Info().Int("count", len(orders)).Msg("orders retrieved")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log(ctx)

	idParam := chi.URLParam(r, "id")
	id, err := validation.GetValidUUID(idParam)
	if err != nil {
		HandleError(ctx, w, err)
		return
	}

	o, err := h.orderService.GetOrderByID(id)

	if err != nil {
		HandleError(ctx, w, err)
		return
	}

	logger.Info().Str("order_id", id.String()).Msg("order was found")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log(ctx)

	var req CreateOrderRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Warn().Err(err).Msg("validation error occurred while creating order")
		HandleError(ctx, w, service.ErrInvalidOrder)
		return
	}

	if err = validateCreateOrder(ctx, req); err != nil {
		HandleError(ctx, w, err)
		return
	}

	logger.Info().Interface("request", req).Msg("create order request received")
	o, err := h.orderService.CreateOrder(r.Context(), req.ProductID, req.Quantity)
	if err != nil {
		HandleError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", "/orders/"+o.ID.String())
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(o)
}

func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log(ctx)

	idParam := chi.URLParam(r, "id")
	id, err := validation.GetValidUUID(idParam)
	if err != nil {
		HandleError(ctx, w, err)
		return
	}
	logger.Info().Str("order_id", id.String()).Msg("update order request received")

	var req UpdateOrderRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Warn().Err(err).Msg("validation error occurred while updating order")
		HandleError(ctx, w, service.ErrInvalidOrder)
		return
	}

	// normalize status - to uppercase
	req.Status = req.Status.Normalize()
	if err = validateUpdateOrder(ctx, req); err != nil {
		HandleError(ctx, w, err)
		return
	}

	logger.Info().Interface("request", req).Msg("update order")
	o, err := h.orderService.UpdateOrder(r.Context(), id, req.ProductID, req.Quantity, req.Status)
	if err != nil {
		HandleError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(o)
}

func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log(ctx)

	idParam := chi.URLParam(r, "id")
	id, err := validation.GetValidUUID(idParam)
	if err != nil {
		HandleError(ctx, w, err)
		return
	}

	logger.Info().Str("order_id", id.String()).Msg("delete order request received")
	h.orderService.DeleteOrder(id)

	w.WriteHeader(http.StatusNoContent)
}
