package service

import (
	"commerce-platform/services/orders/internal/grpc"
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/services/orders/internal/repository"
	"context"
	"testing"

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
	client := &mockProductsClient{productIDs: map[string]bool{"1": true, "2": true}}
	svc := NewOrderService(repo, client)

	return svc, repo, client
}

func TestGetOrders_WhenOrdersExist_ReturnsOrders(t *testing.T) {
	svc, _, _ := setup(t)

	orders := svc.GetOrders()

	assert.Equal(t, 2, len(orders))
}

func TestGetOrderByID_WhenOrderExists_ReturnsOrder(t *testing.T) {
	svc, _, _ := setup(t)

	o, err := svc.GetOrderByID("1")

	assert.NoError(t, err)
	assert.Equal(t, "1", o.ID)
	assert.Equal(t, "1", o.ProductID)
	assert.Equal(t, 2, o.Quantity)
	assert.Equal(t, order.CREATED, o.Status)
}

func TestGetOrderByID_WhenOrderNotExists_ReturnsNotFound(t *testing.T) {
	svc, _, _ := setup(t)

	o, err := svc.GetOrderByID("11")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrOrderNotFound)
	assert.Empty(t, o)
}

func TestCreateOrder_WhenProductExists_CreatesOrder(t *testing.T) {
	svc, repo, _ := setup(t)

	err := svc.CreateOrder(context.Background(), "11", "1", 10)

	assert.NoError(t, err)
	o, exists := repo.FindByID("11")
	assert.True(t, exists)
	assert.Equal(t, order.Order{
		ID:        "11",
		ProductID: "1",
		Quantity:  10,
		Status:    order.CREATED,
	}, o)
}

func TestCreateOrder_WhenProductNotExists_ReturnsError(t *testing.T) {
	svc, repo, _ := setup(t)

	err := svc.CreateOrder(context.Background(), "11", "999", 10)

	assert.ErrorIs(t, err, ErrProductNotFound)
	_, exists := repo.FindByID("11")
	assert.False(t, exists)
}

func TestUpdateOrder_WhenOrderNotExists_CreatesOrder(t *testing.T) {
	svc, repo, _ := setup(t)

	err := svc.UpdateOrder("11", "1", 10, order.CANCELLED)

	assert.Error(t, err)

	o, exists := repo.FindByID("11")

	assert.False(t, exists)
	assert.ErrorIs(t, err, ErrOrderNotFound)
	assert.Empty(t, o)
}

func TestUpdateOrder_WhenOrderExists_UpdatesOrder(t *testing.T) {
	svc, repo, _ := setup(t)

	o, exists := repo.FindByID("1")
	assert.True(t, exists)
	assert.Equal(t, order.Order{
		ID:        "1",
		ProductID: "1",
		Quantity:  2,
		Status:    order.CREATED,
	}, o)

	err := svc.UpdateOrder("1", "1", 11, order.PAID)

	o, exists = repo.FindByID("1")

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, order.Order{
		ID:        "1",
		ProductID: "1",
		Quantity:  11,
		Status:    order.PAID,
	}, o)
}

func TestDeleteOrder_WhenOrderNotExists_DoesNotFail(t *testing.T) {
	svc, repo, _ := setup(t)

	// order does not exist
	_, exists := repo.FindByID("11")
	assert.False(t, exists)

	svc.DeleteOrder("11")
	_, exists = repo.FindByID("11")

	assert.False(t, exists)

	orders := repo.FindAll()
	assert.Len(t, orders, 2)
}

func TestDeleteOrder_WhenOrderExists_DeletesOrder(t *testing.T) {
	svc, repo, _ := setup(t)

	// order exists
	_, exists := repo.FindByID("2")
	assert.True(t, exists)

	svc.DeleteOrder("2")
	_, exists = repo.FindByID("2")

	assert.False(t, exists)

	orders := repo.FindAll()
	assert.Len(t, orders, 1)
}
