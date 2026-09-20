package http

import (
	"commerce-platform/services/orders/internal/service"
	"commerce-platform/services/orders/internal/validation"
	"encoding/json"
	"errors"
	"fmt"
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

	orders, err := h.orderService.GetOrders(ctx)
	if err != nil {
		HandleError(ctx, w, fmt.Errorf("get_orders: %w", err))
		return
	}

	logger.Info().Int("count", len(orders)).Msg("orders retrieved")

	HandleResponseWithBody(ctx, w, http.StatusOK, orders)
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

	o, err := h.orderService.GetOrderByID(ctx, id)
	if err != nil {
		HandleError(ctx, w, fmt.Errorf("get_order: %w", err))
		return
	}

	logger.Info().Str("order_id", id.String()).Msg("order was found")

	HandleResponseWithBody(ctx, w, http.StatusOK, o)
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log(ctx)

	var req CreateOrderRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		err = errors.Join(
			fmt.Errorf("decode create_order request: %w", err),
			service.ErrInvalidOrder,
		)
		HandleError(ctx, w, err)
		return
	}

	if err = validateCreateOrder(ctx, req); err != nil {
		HandleError(ctx, w, err)
		return
	}

	logger.Info().Interface("request", req).Msg("create order request received")
	o, err := h.orderService.CreateOrder(ctx, req.ProductID, req.Quantity)
	if err != nil {
		HandleError(ctx, w, fmt.Errorf("create_order: %w", err))
		return
	}

	HandlePostResponse(ctx, w, http.StatusCreated, o, "/orders/"+o.ID.String())
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
		err = errors.Join(
			fmt.Errorf("decode update_order request: %w", err),
			service.ErrInvalidOrder,
		)
		HandleError(ctx, w, err)
		return
	}

	// normalize status - to uppercase
	req.Status = req.Status.Normalize()
	if err = validateUpdateOrder(ctx, req); err != nil {
		HandleError(ctx, w, err)
		return
	}

	logger.Info().Interface("request", req).Msg("update order")
	o, err := h.orderService.UpdateOrder(ctx, id, req.ProductID, req.Quantity, req.Status)
	if err != nil {
		HandleError(ctx, w, fmt.Errorf("update_order: %w", err))
		return
	}

	HandleResponseWithBody(ctx, w, http.StatusOK, o)
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
	err = h.orderService.DeleteOrder(ctx, id)
	if err != nil {
		HandleError(ctx, w, fmt.Errorf("delete_order: %w", err))
		return
	}

	HandleResponse(ctx, w, http.StatusNoContent)
}
