package http

import "commerce-platform/services/orders/internal/order"

type CreateOrderRequest struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type UpdateOrderRequest struct {
	ProductID string            `json:"product_id"`
	Quantity  int               `json:"quantity"`
	Status    order.OrderStatus `json:"status"`
}
