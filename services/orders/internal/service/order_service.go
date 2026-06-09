package service

import (
	"commerce-platform/services/orders/internal/order"
	"log/slog"
)

type OrderRepository interface {
	FindAll() []order.Order
	FindByID(id string) (order.Order, bool)
	Save(order.Order)
	Update(order.Order)
	Delete(id string)
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

func (s *OrderService) CreateOrder(id string, productID string, quantity int) {
	o := order.Order{
		ID:        id,
		ProductID: productID,
		Quantity:  quantity,
		Status:    order.CREATED,
	}

	slog.Info("creating", "order", o)

	s.repository.Save(o)
}

func (s *OrderService) UpdateOrder(id string, productID string, quantity int, status order.OrderStatus) {
	o := order.Order{
		ID:        id,
		ProductID: productID,
		Quantity:  quantity,
		Status:    status,
	}

	slog.Info("updating", "order", o)

	s.repository.Update(o)
}

func (s *OrderService) DeleteOrder(id string) {
	slog.Info("attempting to delete product with", "productId", id)

	s.repository.Delete(id)
}
