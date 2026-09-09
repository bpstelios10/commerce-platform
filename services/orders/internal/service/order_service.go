package service

import (
	"commerce-platform/services/orders/internal/grpc"
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/shared/logger"
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type OrderRepository interface {
	FindAll(ctx context.Context) []order.Order
	FindByID(ctx context.Context, id uuid.UUID) (order.Order, bool)
	Save(ctx context.Context, o order.Order)
	Update(ctx context.Context, o order.Order)
	Delete(ctx context.Context, id uuid.UUID)
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

func (s *OrderService) GetOrders(ctx context.Context) []order.Order {
	return s.orderRepository.FindAll(ctx)
}

func (s *OrderService) GetOrderByID(ctx context.Context, id uuid.UUID) (order.Order, error) {
	o, found := s.orderRepository.FindByID(ctx, id)

	if !found {
		logger := log(ctx)
		logger.Warn().Str("order_id", id.String()).Msg("order not found")
		return order.Order{}, ErrOrderNotFound
	}

	return o, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, productID string, quantity int) (order.Order, error) {
	if err := s.validateProductExists(ctx, productID); err != nil {
		return order.Order{}, err
	}

	id, _ := uuid.NewV7()

	o := order.Order{
		ID:        id,
		ProductID: productID,
		Quantity:  quantity,
		Status:    order.CREATED,
	}

	logger := log(ctx)
	logger.Info().Str("order_id", o.ID.String()).Str("product_id", o.ProductID).Msg("creating order")

	s.orderRepository.Save(ctx, o)
	return o, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, id uuid.UUID, productID string, quantity int, status order.OrderStatus) (order.Order, error) {
	if _, err := s.GetOrderByID(ctx, id); err != nil {
		return order.Order{}, err
	}

	if err := s.validateProductExists(ctx, productID); err != nil {
		return order.Order{}, err
	}

	o := order.Order{
		ID:        id,
		ProductID: productID,
		Quantity:  quantity,
		Status:    status,
	}

	logger := log(ctx)
	logger.Info().Str("order_id", o.ID.String()).Str("product_id", o.ProductID).Msg("updating order")

	s.orderRepository.Update(ctx, o)
	return o, nil
}

func (s *OrderService) DeleteOrder(ctx context.Context, id uuid.UUID) {
	logger := log(ctx)
	logger.Info().Str("order_id", id.String()).Msg("attempting to delete order")

	s.orderRepository.Delete(ctx, id)
}

// TODO return error. we hide now if it is InvalidArgument, NotFound, Internal
func (s *OrderService) validateProductExists(ctx context.Context, productID string) error {
	_, err := s.productsClient.GetProductByID(ctx, productID)
	if err != nil {
		logger := log(ctx)
		logger.Warn().Str("product_id", productID).Msg("product not found for given product id")
		return ErrProductNotFound
	}
	return nil
}

func log(ctx context.Context) zerolog.Logger {
	return logger.GetLogger(ctx, "orders.service")
}
