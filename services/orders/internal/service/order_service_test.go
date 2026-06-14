package service

import (
	"commerce-platform/services/orders/internal/grpc"
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/services/orders/internal/repository"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// mockProductsClient implements ProductsClient for tests.
type mockProductsClient struct {
	productIDs map[string]bool // product IDs that "exist"
}

func (m *mockProductsClient) GetProductByID(_ context.Context, id string) (*grpc.GetProductByIDResponse, error) {
	if m.productIDs[id] {
		return &grpc.GetProductByIDResponse{Id: id}, nil
	}
	return nil, ErrProductNotFound
}

func setup(t *testing.T) (*OrderService, *repository.InMemoryOrderRepository, *mockProductsClient) {
	t.Helper()
	repo := repository.NewInMemoryOrderRepository()
	client := &mockProductsClient{
		productIDs: map[string]bool{
			repository.FirstProductID:  true,
			repository.SecondProductID: true,
		},
	}
	svc := NewOrderService(repo, client)

	return svc, repo, client
}

func TestGetOrders_WhenOrdersExist_ReturnsOrders(t *testing.T) {
	svc, _, _ := setup(t)

	orders := svc.GetOrders()

	assert.Len(t, orders, 2)
}

func TestGetOrderByID_WhenOrderExists_ReturnsOrder(t *testing.T) {
	svc, _, _ := setup(t)

	o, err := svc.GetOrderByID(repository.FirstOrderID)

	assert.NoError(t, err)
	assert.Equal(t, repository.FirstOrderID, o.ID)
	assert.Equal(t, repository.FirstProductID, o.ProductID)
	assert.Equal(t, 2, o.Quantity)
	assert.Equal(t, order.CREATED, o.Status)
}

func TestGetOrderByID_WhenOrderNotExists_ReturnsNotFound(t *testing.T) {
	svc, _, _ := setup(t)
	id, _ := uuid.NewV7()

	o, err := svc.GetOrderByID(id)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrOrderNotFound)
	assert.Empty(t, o)
}

func TestCreateOrder_WhenProductExists_CreatesOrder(t *testing.T) {
	svc, repo, _ := setup(t)

	o, err := svc.CreateOrder(context.Background(), repository.FirstProductID, 10)

	assert.NoError(t, err)
	o, exists := repo.FindByID(o.ID)
	assert.True(t, exists)
	assert.Equal(t, order.Order{
		ID:        o.ID,
		ProductID: repository.FirstProductID,
		Quantity:  10,
		Status:    order.CREATED,
	}, o)
}

func TestCreateOrder_WhenProductNotExists_ReturnsError(t *testing.T) {
	svc, repo, _ := setup(t)

	o, err := svc.CreateOrder(context.Background(), "999", 10)

	assert.ErrorIs(t, err, ErrProductNotFound)
	assert.Empty(t, o)
	orders := repo.FindAll()
	assert.Len(t, orders, 2)
}

func TestUpdateOrder_WhenOrderNotExists_CreatesOrder(t *testing.T) {
	svc, repo, _ := setup(t)
	id, _ := uuid.NewV7()

	err := svc.UpdateOrder(context.Background(), id, "1", 10, order.CANCELLED)

	assert.Error(t, err)

	o, exists := repo.FindByID(id)

	assert.False(t, exists)
	assert.ErrorIs(t, err, ErrOrderNotFound)
	assert.Empty(t, o)
}

func TestUpdateOrder_WhenOrderExists_UpdatesOrder(t *testing.T) {
	svc, repo, _ := setup(t)

	o, exists := repo.FindByID(repository.FirstOrderID)
	assert.True(t, exists)
	assert.Equal(t, order.Order{
		ID:        repository.FirstOrderID,
		ProductID: repository.FirstProductID,
		Quantity:  2,
		Status:    order.CREATED,
	}, o)

	err := svc.UpdateOrder(context.Background(), repository.FirstOrderID, repository.FirstProductID, 11, order.PAID)

	o, exists = repo.FindByID(repository.FirstOrderID)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, order.Order{
		ID:        repository.FirstOrderID,
		ProductID: repository.FirstProductID,
		Quantity:  11,
		Status:    order.PAID,
	}, o)
}

func TestUpdateOrder_WhenProductNotExists_ReturnsError(t *testing.T) {
	svc, repo, _ := setup(t)

	err := svc.UpdateOrder(context.Background(), repository.FirstOrderID, "999", 11, order.PAID)

	assert.ErrorIs(t, err, ErrProductNotFound)

	// order unchanged
	o, exists := repo.FindByID(repository.FirstOrderID)
	assert.True(t, exists)
	assert.Equal(t, order.Order{
		ID:        repository.FirstOrderID,
		ProductID: repository.FirstProductID,
		Quantity:  2,
		Status:    order.CREATED,
	}, o)
}

func TestDeleteOrder_WhenOrderNotExists_DoesNotFail(t *testing.T) {
	svc, repo, _ := setup(t)
	id, _ := uuid.NewV7()

	// order does not exist
	_, exists := repo.FindByID(id)
	assert.False(t, exists)

	svc.DeleteOrder(id)
	_, exists = repo.FindByID(id)

	assert.False(t, exists)

	orders := repo.FindAll()
	assert.Len(t, orders, 2)
}

func TestDeleteOrder_WhenOrderExists_DeletesOrder(t *testing.T) {
	svc, repo, _ := setup(t)

	// order exists
	_, exists := repo.FindByID(repository.SecondOrderID)
	assert.True(t, exists)

	svc.DeleteOrder(repository.SecondOrderID)
	_, exists = repo.FindByID(repository.SecondOrderID)

	assert.False(t, exists)

	orders := repo.FindAll()
	assert.Len(t, orders, 1)
}
