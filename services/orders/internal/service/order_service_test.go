package service

import (
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/services/orders/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup(t *testing.T) (*OrderService, *repository.InMemoryOrderRepository) {
	t.Helper()
	repo := repository.NewInMemoryOrderRepository()
	svc := NewOrderService(repo)

	return svc, repo
}

func TestGetOrders_WhenOrdersExist_ReturnsOrders(t *testing.T) {
	svc, _ := setup(t)

	orders := svc.GetOrders()

	assert.Equal(t, 2, len(orders))
}

func TestGetOrderByID_WhenOrderExists_ReturnsOrder(t *testing.T) {
	svc, _ := setup(t)

	o, err := svc.GetOrderByID("1")

	assert.NoError(t, err)
	assert.Equal(t, "1", o.ID)
	assert.Equal(t, "1", o.ProductID)
	assert.Equal(t, 2, o.Quantity)
	assert.Equal(t, order.CREATED, o.Status)
}

func TestGetOrderByID_WhenOrderNotExists_ReturnsNotFound(t *testing.T) {
	svc, _ := setup(t)

	_, err := svc.GetOrderByID("11")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrOrderNotFound)
}

func TestCreateOrder_WhenOrderNotExists_CreatesOrder(t *testing.T) {
	svc, repo := setup(t)

	svc.CreateOrder("11", "1", 10)

	o, exists := repo.FindByID("11")

	assert.True(t, exists)
	assert.Equal(t, order.Order{
		ID:        "11",
		ProductID: "1",
		Quantity:  10,
		Status:    order.CREATED,
	}, o)
}

func TestCreateOrder_WhenOrderExists_UpdatesOrder(t *testing.T) {
	svc, repo := setup(t)

	o, err := svc.GetOrderByID("1")
	assert.NoError(t, err)
	assert.Equal(t, "1", o.ID)
	assert.Equal(t, "1", o.ProductID)
	assert.Equal(t, 2, o.Quantity)
	assert.Equal(t, order.CREATED, o.Status)

	svc.CreateOrder("1", "1", 10)

	o, exists := repo.FindByID("1")

	assert.True(t, exists)
	assert.Equal(t, order.Order{
		ID:        "1",
		ProductID: "1",
		Quantity:  10,
		Status:    order.CREATED,
	}, o)
}

func TestUpdateOrder_WhenOrderNotExists_CreatesOrder(t *testing.T) {
	svc, repo := setup(t)

	svc.UpdateOrder("11", "1", 10, order.CANCELLED)

	o, exists := repo.FindByID("11")

	assert.True(t, exists)
	assert.Equal(t, order.Order{
		ID:        "11",
		ProductID: "1",
		Quantity:  10,
		Status:    order.CANCELLED,
	}, o)
}

func TestUpdateOrder_WhenOrderExists_UpdatesOrder(t *testing.T) {
	svc, repo := setup(t)

	o, exists := repo.FindByID("1")
	assert.True(t, exists)
	assert.Equal(t, order.Order{
		ID:        "1",
		ProductID: "1",
		Quantity:  2,
		Status:    order.CREATED,
	}, o)

	svc.UpdateOrder("1", "1", 11, order.PAID)

	o, exists = repo.FindByID("1")

	assert.True(t, exists)
	assert.Equal(t, order.Order{
		ID:        "1",
		ProductID: "1",
		Quantity:  11,
		Status:    order.PAID,
	}, o)
}

func TestDeleteOrder_WhenOrderNotExists_DoesNotFail(t *testing.T) {
	svc, repo := setup(t)

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
	svc, repo := setup(t)

	// order exists
	_, exists := repo.FindByID("2")
	assert.True(t, exists)

	svc.DeleteOrder("2")
	_, exists = repo.FindByID("2")

	assert.False(t, exists)

	orders := repo.FindAll()
	assert.Len(t, orders, 1)
}
