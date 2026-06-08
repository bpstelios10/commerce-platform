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
