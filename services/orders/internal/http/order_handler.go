package http

import (
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/services/orders/internal/service"
	"commerce-platform/services/orders/internal/validation"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OrderService interface {
	GetOrders(ctx context.Context) ([]order.Order, error)
	GetOrderByID(ctx context.Context, id uuid.UUID) (order.Order, error)
	CreateOrder(ctx context.Context, productID string, quantity int) (order.Order, error)
	UpdateOrder(ctx context.Context, id uuid.UUID, productID string, quantity int, status order.OrderStatus) (order.Order, error)
	DeleteOrder(ctx context.Context, id uuid.UUID) error
}

type OrderHandler struct {
	orderService OrderService
}

func NewOrderHandler(service OrderService) *OrderHandler {
	return &OrderHandler{orderService: service}
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
	if !isBodyValidAndDecoded(ctx, r.Body, w, &req, "create_order") {
		return
	}

	// Normalize input
	req.ProductID = strings.TrimSpace(req.ProductID)
	if err := validateCreateOrder(ctx, req); err != nil {
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
	if !isBodyValidAndDecoded(ctx, r.Body, w, &req, "update_order") {
		return
	}

	// normalize input
	req.Status = req.Status.Normalize()
	req.ProductID = strings.TrimSpace(req.ProductID)
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

func isBodyValidAndDecoded[T any](ctx context.Context, body io.ReadCloser, w http.ResponseWriter, req *T, action string) bool {
	body = http.MaxBytesReader(w, body, 1<<10)
	dec := json.NewDecoder(body)

	if err := dec.Decode(&req); err != nil {
		var e *http.MaxBytesError
		if errors.As(err, &e) {
			err = errors.Join(
				fmt.Errorf("decode %s request: %w", action, err),
				TooLongBodyErr{limit: 1 << 10},
			)
		} else {
			err = errors.Join(
				fmt.Errorf("decode %s request: %w", action, err),
				service.ErrInvalidOrder,
			)
		}
		HandleError(ctx, w, err)
		return false
	}

	// trailing JSON/content
	if dec.Decode(&struct{}{}) != io.EOF {
		err := fmt.Errorf("decode %s request: %w", action, service.ErrInvalidOrder)
		HandleError(ctx, w, err)
		return false
	}

	return true
}
