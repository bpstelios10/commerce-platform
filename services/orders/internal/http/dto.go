package http

import "commerce-platform/services/orders/internal/order"

type CreateOrderRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

type UpdateOrderRequest struct {
	ProductID string            `json:"product_id"`
	Quantity  int32             `json:"quantity"`
	Status    order.OrderStatus `json:"status"`
	Version   int               `json:"version"`
}
