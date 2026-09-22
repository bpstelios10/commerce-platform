package service

import (
	"commerce-platform/services/orders/internal/grpc"
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/services/orders/internal/repository"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockProductsClient implements ProductsClient for tests.
type mockProductsClient struct {
	productIDs map[string]bool // product IDs that "exist"
}

func (m *mockProductsClient) GetProductByID(_ context.Context, id string) (*grpc.GetProductByIDResponse, error) {
	if m.productIDs[id] {
		return &grpc.GetProductByIDResponse{Id: id}, nil
	} else if id == "error" {
		return nil, status.Error(codes.Internal, "unexpected error")
	}
	return nil, status.Error(codes.NotFound, "product not found")
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

	orders, err := svc.GetOrders(context.Background())

	assert.NoError(t, err)
	assert.Len(t, orders, 2)
}

func TestGetOrders__WhenOtherError_ReturnsError(t *testing.T) {
	svc, _, _ := setup(t)
	ctx := context.Background()
	ctxWithError := context.WithValue(ctx, "errorEnabler", "unexpected error")

	orders, err := svc.GetOrders(ctxWithError)

	assert.Error(t, err)
	assert.EqualError(t, err, "get orders: unexpected error")
	assert.Empty(t, orders)
}

func TestGetOrderByID_WhenOrderExists_ReturnsOrder(t *testing.T) {
	svc, _, _ := setup(t)

	o, err := svc.GetOrderByID(context.Background(), repository.FirstOrderID)

	assert.NoError(t, err)
	assert.Equal(t, repository.FirstOrderID, o.ID)
	assert.Equal(t, repository.FirstProductID, o.ProductID)
	assert.Equal(t, 2, o.Quantity)
	assert.Equal(t, order.CREATED, o.Status)
}

func TestGetOrderByID_WhenOrderNotExists_ReturnsNotFound(t *testing.T) {
	svc, _, _ := setup(t)
	id, _ := uuid.NewV7()

	o, err := svc.GetOrderByID(context.Background(), id)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrOrderNotFound)
	assert.Empty(t, o)
}

func TestGetOrderByID_WhenOtherError_ReturnsError(t *testing.T) {
	svc, _, _ := setup(t)

	o, err := svc.GetOrderByID(context.Background(), repository.ErrornousUUID)

	assert.Error(t, err)
	assert.EqualError(t, err, "get order by id: unexpected error")
	assert.Empty(t, o)
}

func TestCreateOrder_WhenProductExists_CreatesOrder(t *testing.T) {
	svc, repo, _ := setup(t)

	o, err := svc.CreateOrder(context.Background(), repository.FirstProductID, 10)

	assert.NoError(t, err)
	o, err = repo.FindByID(context.Background(), o.ID)
	assert.NoError(t, err)
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
	orders, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, orders, 2)
}

func TestCreateOrder_WhenProductValidationFails_ReturnsError(t *testing.T) {
	svc, repo, _ := setup(t)

	o, err := svc.CreateOrder(context.Background(), "error", 10)

	assert.Error(t, err)
	assert.Empty(t, o)
	orders, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, orders, 2)
}

func TestCreateOrder_WhenDbError_ReturnsError(t *testing.T) {
	svc, repo, _ := setup(t)
	ctx := context.Background()
	ctxWithError := context.WithValue(ctx, "errorEnabler", "unexpected error")

	o, err := svc.CreateOrder(ctxWithError, repository.FirstProductID, 10)

	assert.Error(t, err)
	assert.EqualError(t, err, "create order: unexpected error")
	assert.Empty(t, o)
	orders, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, orders, 2)
}

func TestUpdateOrder_WhenOrderNotExists_CreatesOrder(t *testing.T) {
	svc, repo, _ := setup(t)
	id, _ := uuid.NewV7()

	updated, err := svc.UpdateOrder(context.Background(), id, "1", 10, order.CANCELED)

	assert.Error(t, err)

	o, dbErr := repo.FindByID(context.Background(), id)

	assert.ErrorIs(t, dbErr, repository.ErrNotFound)
	assert.ErrorIs(t, err, ErrOrderNotFound)
	assert.Empty(t, o)
	assert.Empty(t, updated)
}

func TestUpdateOrder_WhenOrderExists_UpdatesOrder(t *testing.T) {
	svc, repo, _ := setup(t)

	o, err := repo.FindByID(context.Background(), repository.FirstOrderID)
	assert.NoError(t, err)
	assert.Equal(t, order.Order{
		ID:        repository.FirstOrderID,
		ProductID: repository.FirstProductID,
		Quantity:  2,
		Status:    order.CREATED,
	}, o)

	updated, err := svc.UpdateOrder(context.Background(), repository.FirstOrderID, repository.FirstProductID, 11, order.PAID)

	o, err = repo.FindByID(context.Background(), repository.FirstOrderID)

	assert.NoError(t, err)
	assert.Equal(t, order.Order{
		ID:        repository.FirstOrderID,
		ProductID: repository.FirstProductID,
		Quantity:  11,
		Status:    order.PAID,
	}, updated)
	assert.NoError(t, err)
	assert.Equal(t, order.Order{
		ID:        repository.FirstOrderID,
		ProductID: repository.FirstProductID,
		Quantity:  11,
		Status:    order.PAID,
	}, o)
}

func TestUpdateOrder_WhenProductNotExists_ReturnsError(t *testing.T) {
	svc, repo, _ := setup(t)

	updated, err := svc.UpdateOrder(context.Background(), repository.FirstOrderID, "999", 11, order.PAID)

	assert.ErrorIs(t, err, ErrProductNotFound)
	assert.Empty(t, updated)

	// order unchanged
	o, err := repo.FindByID(context.Background(), repository.FirstOrderID)
	assert.NoError(t, err)
	assert.Equal(t, order.Order{
		ID:        repository.FirstOrderID,
		ProductID: repository.FirstProductID,
		Quantity:  2,
		Status:    order.CREATED,
	}, o)
}

func TestUpdateOrder_WhenProductValidationFails_ReturnsError(t *testing.T) {
	svc, repo, _ := setup(t)

	updated, err := svc.UpdateOrder(context.Background(), repository.FirstOrderID, "error", 11, order.PAID)

	assert.Error(t, err)
	assert.Empty(t, updated)

	// order unchanged
	o, err := repo.FindByID(context.Background(), repository.FirstOrderID)
	assert.NoError(t, err)
	assert.Equal(t, order.Order{
		ID:        repository.FirstOrderID,
		ProductID: repository.FirstProductID,
		Quantity:  2,
		Status:    order.CREATED,
	}, o)
}

func TestUpdateOrder_WhenDbError_ReturnsError(t *testing.T) {
	svc, repo, _ := setup(t)
	ctx := context.Background()
	ctxWithError := context.WithValue(ctx, "errorEnabler", "unexpected error")

	updated, err := svc.UpdateOrder(ctxWithError, repository.FirstOrderID, repository.FirstProductID, 10, order.CANCELED)

	assert.Error(t, err)
	assert.EqualError(t, err, "update order: unexpected error")
	assert.Empty(t, updated)
	orders, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, orders, 2)
}

func TestDeleteOrder_WhenOrderNotExists_DoesNotFail(t *testing.T) {
	svc, repo, _ := setup(t)
	id, _ := uuid.NewV7()

	// order does not exist
	_, err := repo.FindByID(context.Background(), id)
	assert.ErrorIs(t, err, repository.ErrNotFound)

	svc.DeleteOrder(context.Background(), id)
	_, err = repo.FindByID(context.Background(), id)

	assert.ErrorIs(t, err, repository.ErrNotFound)

	orders, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, orders, 2)
}

func TestDeleteOrder_WhenOrderExists_DeletesOrder(t *testing.T) {
	svc, repo, _ := setup(t)

	// order exists
	_, err := repo.FindByID(context.Background(), repository.SecondOrderID)
	assert.NoError(t, err)

	svc.DeleteOrder(context.Background(), repository.SecondOrderID)
	_, err = repo.FindByID(context.Background(), repository.SecondOrderID)

	assert.ErrorIs(t, err, repository.ErrNotFound)

	orders, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, orders, 1)
}

func TestDeleteOrder_WhenDbError_ReturnsError(t *testing.T) {
	svc, repo, _ := setup(t)
	ctx := context.Background()
	ctxWithError := context.WithValue(ctx, "errorEnabler", "unexpected error")

	err := svc.DeleteOrder(ctxWithError, repository.SecondOrderID)
	assert.Error(t, err)

	_, err = repo.FindByID(context.Background(), repository.SecondOrderID)
	assert.NoError(t, err) // The order should still exist because the delete failed

	orders, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, orders, 2) // The number of orders should remain unchanged
}
