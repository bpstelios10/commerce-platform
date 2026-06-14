package service

import (
	"commerce-platform/services/orders/internal/grpc"
	"commerce-platform/services/orders/internal/order"
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type OrderRepository interface {
	FindAll() []order.Order
	FindByID(id uuid.UUID) (order.Order, bool)
	Save(order.Order)
	Update(order.Order)
	Delete(id uuid.UUID)
}

type ProductsClient interface {
	GetProductByID(ctx context.Context, id string) (*grpc.GetProductByIDResponse, error)
}

type OrderService struct {
	orderRepository OrderRepository
	productsClient  ProductsClient
}

func NewOrderService(repository OrderRepository, productsClient ProductsClient) *OrderService {
	return &OrderService{orderRepository: repository, productsClient: productsClient}
}

func (s *OrderService) GetOrders() []order.Order {
	return s.orderRepository.FindAll()
}

func (s *OrderService) GetOrderByID(id uuid.UUID) (order.Order, error) {
	o, found := s.orderRepository.FindByID(id)

	if !found {
		slog.Warn("Order Not found, with", "ID", id)
		return order.Order{}, ErrOrderNotFound
	}

	return o, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, id uuid.UUID, productID string, quantity int) error {
	if err := s.validateProductExists(ctx, productID); err != nil {
		return err
	}

	o := order.Order{
		ID:        id,
		ProductID: productID,
		Quantity:  quantity,
		Status:    order.CREATED,
	}

	slog.Info("creating", "order", o)

	s.orderRepository.Save(o)
	return nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, id uuid.UUID, productID string, quantity int, status order.OrderStatus) error {
	if _, err := s.GetOrderByID(id); err != nil {
		return err
	}

	if err := s.validateProductExists(ctx, productID); err != nil {
		return err
	}

	o := order.Order{
		ID:        id,
		ProductID: productID,
		Quantity:  quantity,
		Status:    status,
	}

	slog.Info("updating", "order", o)

	s.orderRepository.Update(o)
	return nil
}

func (s *OrderService) DeleteOrder(id uuid.UUID) {
	slog.Info("attempting to delete order with", "productId", id)

	s.orderRepository.Delete(id)
}

// TODO return error. we hide now if it is InvalidArgument, NotFound, Internal
func (s *OrderService) validateProductExists(ctx context.Context, productID string) error {
	_, err := s.productsClient.GetProductByID(ctx, productID)
	if err != nil {
		slog.Warn("product not found for given product id", "productId", productID)
		return ErrProductNotFound
	}
	return nil
}
