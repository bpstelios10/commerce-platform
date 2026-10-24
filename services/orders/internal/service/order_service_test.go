package service

import (
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/services/orders/internal/repository"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (*OrderServiceImpl, *fakeOrderRepo, *fakeProductsClient) {
	t.Helper()
	fakeRepo := newFakeOrderRepo(t)
	client := &fakeProductsClient{
		productIDs: map[string]bool{
			FirstProductID:  true,
			SecondProductID: true,
		},
	}
	svc := NewOrderService(fakeRepo, client)

	return svc, fakeRepo, client
}

func TestGetOrders_WhenOrdersExist_ReturnsOrders(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.expectList(
		func(_ context.Context) ([]order.Order, error) {
			return []order.Order{
				testO,
				{
					ID:        SecondOrderID,
					ProductID: SecondProductID,
					Quantity:  1,
					Status:    order.PAID,
					CreatedAt: time.Now(),
				},
			}, nil
		})

	orders, err := svc.GetOrders(context.Background())

	assert.NoError(t, err)
	assert.Len(t, orders, 2)
}

func TestGetOrders_WhenOtherError_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.expectList(
		func(_ context.Context) ([]order.Order, error) {
			return nil, errors.New("query orders: database unavailable")
		})

	orders, err := svc.GetOrders(context.Background())

	assert.Error(t, err)
	assert.EqualError(t, err, "get orders: query orders: database unavailable")
	assert.Empty(t, orders)
}

func TestGetOrderByID_WhenOrderExists_ReturnsOrder(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.expectGetByID(
		func(_ context.Context, id uuid.UUID) (order.Order, error) {
			return testO, nil
		})

	o, err := svc.GetOrderByID(context.Background(), testO.ID)

	assert.NoError(t, err)
	assert.Equal(t, testO, o)
	assert.Equal(t, testO.ID, mock.searchedIDs[0])
}

func TestGetOrderByID_WhenOrderNotExists_ReturnsNotFound(t *testing.T) {
	svc, mock, _ := setup(t)
	id, _ := uuid.NewV7()
	mock.expectGetByID(
		func(_ context.Context, _ uuid.UUID) (order.Order, error) {
			return order.Order{}, fmt.Errorf("blah: %w", repository.ErrNotFound)
		})

	o, err := svc.GetOrderByID(context.Background(), id)

	assert.Error(t, err)
	assert.EqualError(t, err, "get order by id: Order with id ["+id.String()+"] was not found")
	var notFoundErr *ErrOrderNotFound
	assert.ErrorAs(t, err, &notFoundErr)
	assert.Equal(t, id, notFoundErr.OrderID)
	assert.Empty(t, o)
	assert.Equal(t, id, mock.searchedIDs[0])
}

func TestGetOrderByID_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	id, _ := uuid.NewV7()
	mock.expectGetByID(
		func(_ context.Context, _ uuid.UUID) (order.Order, error) {
			return order.Order{}, errors.New("find order by id: database unavailable")
		})

	o, err := svc.GetOrderByID(context.Background(), id)

	assert.Error(t, err)
	assert.EqualError(t, err, "get order by id: find order by id: database unavailable")
	assert.Empty(t, o)
	assert.Equal(t, id, mock.searchedIDs[0])
}

func TestCreateOrder_WhenProductExists_CreatesOrder(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.expectCreate(
		func(_ context.Context, o order.Order) (order.Order, error) {
			o.CreatedAt = time.Now() // the database fills this in
			return o, nil
		})

	o, err := svc.CreateOrder(context.Background(), FirstProductID, 10)

	require.NoError(t, err)
	assert.NotEmpty(t, o)
	assert.WithinDuration(t, time.Now(), o.CreatedAt, time.Second)
	assert.Equal(t, o.ID, mock.created[0].ID)
	assert.Equal(t, FirstProductID, mock.created[0].ProductID)
	assert.Equal(t, int32(10), mock.created[0].Quantity)
	assert.Equal(t, order.CREATED, mock.created[0].Status)
}

func TestCreateOrder_WhenProductNotExists_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)

	o, err := svc.CreateOrder(context.Background(), "999", 10)

	assert.ErrorIs(t, err, ErrProductNotFound)
	assert.Empty(t, o)
	assert.Empty(t, mock.created)
}

type failingReader struct {
	err error
}

func (r failingReader) Read([]byte) (int, error) {
	return 0, r.err
}

func TestCreateOrder_WhenUUIDGenerationFails_ReturnsError(t *testing.T) {
	entropyErr := errors.New("entropy unavailable")
	uuid.SetRand(failingReader{err: entropyErr})
	t.Cleanup(func() { uuid.SetRand(nil) }) // resets to crypto/rand.Reader
	svc, mock, _ := setup(t)

	_, err := svc.CreateOrder(context.Background(), FirstProductID, 10)

	assert.ErrorIs(t, err, entropyErr)
	assert.EqualError(t, err, "create order: entropy unavailable")
	assert.Empty(t, mock.created)
}

func TestCreateOrder_WhenProductValidationFails_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)

	o, err := svc.CreateOrder(context.Background(), "error", 10)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrProductNotFound)
	assert.EqualError(t, err, "create order: get product error from products service: rpc error: code = Internal desc = unexpected error")
	assert.Empty(t, o)
	assert.Empty(t, mock.created)
}

func TestCreateOrder_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.expectCreate(
		func(_ context.Context, o order.Order) (order.Order, error) {
			return o, errors.New("save order: database unavailable")
		})

	o, err := svc.CreateOrder(context.Background(), FirstProductID, 10)

	assert.Error(t, err)
	assert.EqualError(t, err, "create order: save order: database unavailable")
	assert.Empty(t, o)
	assert.NotEqual(t, uuid.Nil, mock.created[0].ID)
	assert.Equal(t, FirstProductID, mock.created[0].ProductID)
	assert.Equal(t, int32(10), mock.created[0].Quantity)
	assert.Equal(t, order.CREATED, mock.created[0].Status)
}

func TestUpdateOrder_WhenOrderExists_UpdatesOrder(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.expectGetByID(
		func(_ context.Context, id uuid.UUID) (order.Order, error) {
			return testO, nil
		})
	mock.expectUpdate(
		func(_ context.Context, o order.Order) (order.Order, error) {
			return o, nil
		})

	updated, err := svc.UpdateOrder(context.Background(), testO.ID, FirstProductID, 11, order.PAID, 1)

	assert.NoError(t, err)
	assert.Equal(t, testO.ID, updated.ID)
	assert.Equal(t, FirstProductID, updated.ProductID)
	assert.Equal(t, int32(11), updated.Quantity)
	assert.Equal(t, order.PAID, updated.Status)
	assert.True(t, testO.CreatedAt.Equal(updated.CreatedAt))
	assert.Equal(t, 2, updated.Version)

	assert.Equal(t, testO.ID, mock.updated[0].ID)
	assert.Equal(t, FirstProductID, mock.updated[0].ProductID)
	assert.Equal(t, int32(11), mock.updated[0].Quantity)
	assert.Equal(t, order.PAID, mock.updated[0].Status)
	assert.Equal(t, 1, mock.updated[0].Version)
}

func TestUpdateOrder_WhenOrderNotExists_CreatesOrder(t *testing.T) {
	svc, mock, _ := setup(t)
	id := uuid.New()
	mock.expectGetByID(
		func(_ context.Context, _ uuid.UUID) (order.Order, error) {
			return order.Order{}, fmt.Errorf("blah: %w", repository.ErrNotFound)
		})

	updated, err := svc.UpdateOrder(context.Background(), id, "1", 10, order.CANCELED, 1)

	assert.Error(t, err)
	err1 := &ErrOrderNotFound{}
	assert.True(t, errors.As(err, &err1))
	assert.Equal(t, id, err1.OrderID)
	assert.EqualError(t, err, "updating order: get order by id: Order with id ["+id.String()+"] was not found")
	assert.Empty(t, updated)
	assert.Equal(t, id, mock.searchedIDs[0])
	assert.Empty(t, mock.updated)
}

func TestUpdateOrder_WhenProductNotExists_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.expectGetByID(
		func(_ context.Context, id uuid.UUID) (order.Order, error) {
			o := testO
			o.CreatedAt = time.Now()
			return o, nil
		})

	updated, err := svc.UpdateOrder(context.Background(), FirstOrderID, "999", 11, order.PAID, 1)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrProductNotFound)
	assert.EqualError(t, err, "updating order: get product 999 from products service: rpc error: code = NotFound desc = product not found\nproduct not found for the given id")
	assert.Empty(t, updated)
	assert.Equal(t, testO.ID, mock.searchedIDs[0])
	assert.Empty(t, mock.updated)
}

func TestUpdateOrder_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.expectGetByID(
		func(_ context.Context, id uuid.UUID) (order.Order, error) {
			return testO, nil
		})
	mock.expectUpdate(
		func(_ context.Context, o order.Order) (order.Order, error) {
			return o, errors.New("update order: database unavailable")
		})

	updated, err := svc.UpdateOrder(context.Background(), testO.ID, FirstProductID, 10, order.CANCELED, 1)

	assert.Error(t, err)
	assert.EqualError(t, err, "updating order: update order: database unavailable")
	assert.Empty(t, updated)
	assert.Equal(t, testO.ID, mock.searchedIDs[0])
	assert.Equal(t, testO.ID, mock.updated[0].ID)
	assert.Equal(t, FirstProductID, mock.updated[0].ProductID)
	assert.Equal(t, int32(10), mock.updated[0].Quantity)
	assert.Equal(t, order.CANCELED, mock.updated[0].Status)
}

func TestUpdateOrder_WhenUpdateConflict_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.expectGetByID(
		func(_ context.Context, id uuid.UUID) (order.Order, error) {
			return testO, nil
		})
	mock.expectUpdate(
		func(_ context.Context, o order.Order) (order.Order, error) {
			return o, fmt.Errorf("blah: %w", repository.ErrVersionConflict)
		})

	updated, err := svc.UpdateOrder(context.Background(), testO.ID, FirstProductID, 10, order.CANCELED, 1)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrOrderUpdateConflict)
	assert.EqualError(t, err, "updating order: order was updated by another process")
	assert.Empty(t, updated)
	assert.Equal(t, testO.ID, mock.searchedIDs[0])
	assert.Equal(t, testO.ID, mock.updated[0].ID)
	assert.Equal(t, FirstProductID, mock.updated[0].ProductID)
	assert.Equal(t, int32(10), mock.updated[0].Quantity)
	assert.Equal(t, order.CANCELED, mock.updated[0].Status)
}

func TestDeleteOrder_WhenOrderNotExists_DoesNotFail(t *testing.T) {
	svc, mock, _ := setup(t)
	toDelete, _ := uuid.NewV7()
	mock.expectDelete(
		func(_ context.Context, id uuid.UUID) error {
			return nil
		})

	err := svc.DeleteOrder(context.Background(), toDelete)

	assert.NoError(t, err)
	assert.Equal(t, toDelete, mock.deleted[0])
}

func TestDeleteOrder_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	toDelete, _ := uuid.NewV7()
	mock.expectDelete(
		func(_ context.Context, id uuid.UUID) error {
			return errors.New("delete order: database unavailable")
		})

	err := svc.DeleteOrder(context.Background(), toDelete)

	assert.Error(t, err)
	assert.EqualError(t, err, "deleting order: delete order: database unavailable")
	assert.Equal(t, toDelete, mock.deleted[0])
}

var (
	FirstProductID  = "f47ac10b-58cc-4372-a567-0e02b2c3d001"
	SecondProductID = "f47ac10b-58cc-4372-a567-0e02b2c3d002"

	FirstOrderID  = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d011")
	SecondOrderID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d012")

	testO = order.Order{
		ID:        FirstOrderID,
		ProductID: FirstProductID,
		Quantity:  int32(2),
		Status:    order.CREATED,
		CreatedAt: time.Now(),
		Version:   1,
	}
)
