package service

import (
	"commerce-platform/services/orders/internal/grpc"
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/services/orders/internal/repository"
	"commerce-platform/shared/logger"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderRepository interface {
	FindAll(ctx context.Context) ([]order.Order, error)
	FindByID(ctx context.Context, id uuid.UUID) (order.Order, error)
	Save(ctx context.Context, o order.Order) error
	Update(ctx context.Context, o order.Order) error
	Delete(ctx context.Context, id uuid.UUID) error
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

func (s *OrderService) GetOrders(ctx context.Context) ([]order.Order, error) {
	orders, err := s.orderRepository.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get orders: %w", err)
	}

	return orders, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, id uuid.UUID) (order.Order, error) {
	o, err := s.orderRepository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return order.Order{}, &ErrOrderNotFound{OrderID: id}
		}

		return order.Order{}, fmt.Errorf("get order by id: %w", err)
	}

	return o, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, productID string, quantity int) (order.Order, error) {
	if err := s.validateProductExists(ctx, productID); err != nil {
		return order.Order{}, fmt.Errorf("create order: %w", err)
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

	err := s.orderRepository.Save(ctx, o)
	if err != nil {
		return order.Order{}, fmt.Errorf("create order: %w", err)
	}

	return o, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, id uuid.UUID, productID string, quantity int, status order.OrderStatus) (order.Order, error) {
	if _, err := s.GetOrderByID(ctx, id); err != nil {
		return order.Order{}, fmt.Errorf("update order: %w", err)
	}

	if err := s.validateProductExists(ctx, productID); err != nil {
		return order.Order{}, fmt.Errorf("update order: %w", err)
	}

	o := order.Order{
		ID:        id,
		ProductID: productID,
		Quantity:  quantity,
		Status:    status,
	}

	logger := log(ctx)
	logger.Info().Str("order_id", o.ID.String()).Str("product_id", o.ProductID).Msg("updating order")

	err := s.orderRepository.Update(ctx, o)
	if err != nil {
		return order.Order{}, fmt.Errorf("update order: %w", err)
	}
	return o, nil
}

func (s *OrderService) DeleteOrder(ctx context.Context, id uuid.UUID) error {
	logger := log(ctx)
	logger.Info().Str("order_id", id.String()).Msg("attempting to delete order")

	err := s.orderRepository.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("delete order: %w", err)
	}
	return nil
}

func (s *OrderService) validateProductExists(ctx context.Context, productID string) error {
	_, err := s.productsClient.GetProductByID(ctx, productID)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return errors.Join(
				fmt.Errorf("get product %s from products service: %w", productID, err),
				ErrProductNotFound,
			)
		}
		return fmt.Errorf("get product %s from products service: %w", productID, err)
	}
	return nil
}

func log(ctx context.Context) zerolog.Logger {
	return logger.GetLogger(ctx, "orders.service")
}
