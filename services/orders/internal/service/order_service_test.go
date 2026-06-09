package service

import (
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/services/orders/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOrders_WhenOrdersExist_ReturnsOrders(t *testing.T) {
	repo := repository.NewInMemoryOrderRepository()
	svc := NewOrderService(repo)

	orders := svc.GetOrders()

	assert.Equal(t, 2, len(orders))
}

func TestGetOrderByID_WhenOrderExists_ReturnsOrder(t *testing.T) {
	repo := repository.NewInMemoryOrderRepository()
	svc := NewOrderService(repo)

	o, err := svc.GetOrderByID("1")

	assert.Nil(t, err)
	assert.Equal(t, "1", o.ID)
	assert.Equal(t, "1", o.ProductID)
	assert.Equal(t, 2, o.Quantity)
	assert.Equal(t, order.CREATED, o.Status)
}

func TestGetOrderByID_WhenOrderNotExists_ReturnsNotFound(t *testing.T) {
	repo := repository.NewInMemoryOrderRepository()
	svc := NewOrderService(repo)

	_, err := svc.GetOrderByID("11")

	assert.NotNil(t, err)
	assert.Equal(t, ErrOrderNotFound, err)
}

func TestCreateOrder_WhenOrderNotExists_CreatesOrder(t *testing.T) {
	repo := repository.NewInMemoryOrderRepository()
	svc := NewOrderService(repo)

	svc.CreateOrder("11", "1", 10)

	o, err := svc.GetOrderByID("11")

	assert.Nil(t, err)
	assert.Equal(t, "11", o.ID)
	assert.Equal(t, "1", o.ProductID)
	assert.Equal(t, 10, o.Quantity)
	assert.Equal(t, order.CREATED, o.Status)
}

func TestCreateOrder_WhenOrderExists_UpdatesOrder(t *testing.T) {
	repo := repository.NewInMemoryOrderRepository()
	svc := NewOrderService(repo)

	o, err := svc.GetOrderByID("1")
	assert.Nil(t, err)
	assert.Equal(t, "1", o.ID)
	assert.Equal(t, "1", o.ProductID)
	assert.Equal(t, 2, o.Quantity)
	assert.Equal(t, order.CREATED, o.Status)

	svc.CreateOrder("1", "1", 10)

	o, err = svc.GetOrderByID("1")

	assert.Nil(t, err)
	assert.Equal(t, "1", o.ID)
	assert.Equal(t, "1", o.ProductID)
	assert.Equal(t, 10, o.Quantity)
	assert.Equal(t, order.CREATED, o.Status)
}

func TestUpdateOrder_WhenOrderNotExists_CreatesOrder(t *testing.T) {
	repo := repository.NewInMemoryOrderRepository()
	svc := NewOrderService(repo)

	svc.UpdateOrder("11", "1", 10, order.CANCELLED)

	o, err := svc.GetOrderByID("11")

	assert.Nil(t, err)
	assert.Equal(t, "11", o.ID)
	assert.Equal(t, "1", o.ProductID)
	assert.Equal(t, 10, o.Quantity)
	assert.Equal(t, order.CANCELLED, o.Status)
}

func TestUpdateOrder_WhenOrderExists_UpdatesOrder(t *testing.T) {
	repo := repository.NewInMemoryOrderRepository()
	svc := NewOrderService(repo)

	o, err := svc.GetOrderByID("1")
	assert.Nil(t, err)
	assert.Equal(t, "1", o.ID)
	assert.Equal(t, "1", o.ProductID)
	assert.Equal(t, 2, o.Quantity)
	assert.Equal(t, order.CREATED, o.Status)

	svc.UpdateOrder("1", "1", 11, order.PAID)

	o, err = svc.GetOrderByID("1")

	assert.Nil(t, err)
	assert.Equal(t, "1", o.ID)
	assert.Equal(t, "1", o.ProductID)
	assert.Equal(t, 11, o.Quantity)
	assert.Equal(t, order.PAID, o.Status)
}

func TestDeleteOrder_WhenOrderNotExists_DoesNotFail(t *testing.T) {
	repo := repository.NewInMemoryOrderRepository()
	svc := NewOrderService(repo)

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
	repo := repository.NewInMemoryOrderRepository()
	svc := NewOrderService(repo)

	// order exists
	_, exists := repo.FindByID("2")
	assert.True(t, exists)

	svc.DeleteOrder("2")
	_, exists = repo.FindByID("2")

	assert.False(t, exists)

	orders := repo.FindAll()
	assert.Len(t, orders, 1)
}
