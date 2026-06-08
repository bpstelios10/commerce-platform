package service

import (
	"commerce-platform/services/orders/internal/order"
	"log/slog"
)

type OrderRepository interface {
	FindAll() []order.Order
	FindByID(id string) (order.Order, bool)
}

type OrderService struct {
	repository OrderRepository
}

func NewOrderService(repository OrderRepository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) GetOrders() []order.Order {
	return s.repository.FindAll()
}

func (s *OrderService) GetOrderByID(id string) (order.Order, error) {
	o, found := s.repository.FindByID(id)

	if !found {
		slog.Warn("Order Not found, with", "ID", id)
		return order.Order{}, ErrOrderNotFound
	}

	return o, nil
}
