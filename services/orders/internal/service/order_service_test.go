package service

import (
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/services/orders/internal/repository"
	"context"
	"errors"
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
	mock.expectFindAll(
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
	mock.expectFindAll(
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
	mock.expectFindByID(
		func(_ context.Context, id uuid.UUID) (order.Order, error) {
			if id == testO.ID {
				return testO, nil
			} else {
				return order.Order{}, errors.New("wrong id passed in fake")
			}
		})

	o, err := svc.GetOrderByID(context.Background(), testO.ID)

	assert.NoError(t, err)
	assert.Equal(t, testO, o)
}

func TestGetOrderByID_WhenOrderNotExists_ReturnsNotFound(t *testing.T) {
	svc, mock, _ := setup(t)
	id, _ := uuid.NewV7()
	mock.expectFindByID(
		func(_ context.Context, _ uuid.UUID) (order.Order, error) {
			return order.Order{}, repository.ErrNotFound
		})

	o, err := svc.GetOrderByID(context.Background(), id)

	assert.Error(t, err)
	assert.EqualError(t, err, "Order with id ["+id.String()+"] was not found")
	var notFoundErr *ErrOrderNotFound
	assert.ErrorAs(t, err, &notFoundErr)
	assert.Equal(t, id, notFoundErr.OrderID)
	assert.Empty(t, o)
}

func TestGetOrderByID_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	id, _ := uuid.NewV7()
	mock.expectFindByID(
		func(_ context.Context, _ uuid.UUID) (order.Order, error) {
			return order.Order{}, errors.New("find order by id: database unavailable")
		})

	o, err := svc.GetOrderByID(context.Background(), id)

	assert.Error(t, err)
	assert.EqualError(t, err, "get order by id: find order by id: database unavailable")
	assert.Empty(t, o)
}

func TestCreateOrder_WhenProductExists_CreatesOrder(t *testing.T) {
	svc, mockRepo, _ := setup(t)
	mockRepo.expectSave(
		func(_ context.Context, o order.Order) (order.Order, error) {
			o.CreatedAt = time.Now() // the database fills this in
			return o, nil
		})

	o, err := svc.CreateOrder(context.Background(), FirstProductID, 10)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, mockRepo.saved[0].ID)
	assert.Equal(t, FirstProductID, mockRepo.saved[0].ProductID)
	assert.Equal(t, 10, mockRepo.saved[0].Quantity)
	assert.Equal(t, order.CREATED, mockRepo.saved[0].Status)
	assert.Equal(t, mockRepo.saved[0].ID, o.ID)
	assert.WithinDuration(t, time.Now(), o.CreatedAt, time.Second)
}

func TestCreateOrder_WhenProductNotExists_ReturnsError(t *testing.T) {
	svc, _, _ := setup(t)

	o, err := svc.CreateOrder(context.Background(), "999", 10)

	assert.ErrorIs(t, err, ErrProductNotFound)
	assert.Empty(t, o)
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
	svc, _, _ := setup(t)

	_, err := svc.CreateOrder(context.Background(), FirstProductID, 10)

	assert.ErrorIs(t, err, entropyErr)
	assert.ErrorIs(t, err, entropyErr)
	assert.ErrorIs(t, err, entropyErr)
	assert.EqualError(t, err, "create order: entropy unavailable")
}

func TestCreateOrder_WhenProductValidationFails_ReturnsError(t *testing.T) {
	svc, _, _ := setup(t)

	o, err := svc.CreateOrder(context.Background(), "error", 10)

	assert.Error(t, err)
	assert.EqualError(t, err, "create order: get product error from products service: rpc error: code = Internal desc = unexpected error")
	assert.Empty(t, o)
}

func TestCreateOrder_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.expectSave(
		func(_ context.Context, o order.Order) (order.Order, error) {
			return o, errors.New("save order: database unavailable")
		})

	o, err := svc.CreateOrder(context.Background(), FirstProductID, 10)

	assert.Error(t, err)
	assert.EqualError(t, err, "create order: save order: database unavailable")
	assert.Empty(t, o)
}

func TestUpdateOrder_WhenOrderExists_UpdatesOrder(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.expectFindByID(
		func(_ context.Context, id uuid.UUID) (order.Order, error) {
			if id == testO.ID {
				return testO, nil
			} else {
				return order.Order{}, errors.New("wrong id passed in fake")
			}
		})
	mock.expectUpdate(
		func(_ context.Context, o order.Order) (order.Order, error) {
			return o, nil
		})

	updated, err := svc.UpdateOrder(context.Background(), testO.ID, FirstProductID, 11, order.PAID)

	assert.NoError(t, err)
	assert.Equal(t, testO.ID, updated.ID)
	assert.Equal(t, FirstProductID, updated.ProductID)
	assert.Equal(t, 11, updated.Quantity)
	assert.Equal(t, order.PAID, updated.Status)
	assert.True(t, testO.CreatedAt.Equal(updated.CreatedAt))
}

func TestUpdateOrder_WhenOrderNotExists_CreatesOrder(t *testing.T) {
	svc, mock, _ := setup(t)
	id := uuid.New()
	mock.expectFindByID(
		func(_ context.Context, _ uuid.UUID) (order.Order, error) {
			return order.Order{}, repository.ErrNotFound
		})

	updated, err := svc.UpdateOrder(context.Background(), id, "1", 10, order.CANCELED)

	assert.Error(t, err)
	assert.EqualError(t, err, "updating order: Order with id ["+id.String()+"] was not found")
	assert.Empty(t, updated)
}

func TestUpdateOrder_WhenProductNotExists_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.expectFindByID(
		func(_ context.Context, id uuid.UUID) (order.Order, error) {
			if id == testO.ID {
				o := testO
				o.CreatedAt = time.Now()
				return o, nil
			} else {
				return order.Order{}, errors.New("wrong id passed in fake")
			}
		})

	updated, err := svc.UpdateOrder(context.Background(), FirstOrderID, "999", 11, order.PAID)

	assert.Error(t, err)
	assert.EqualError(t, err, "updating order: get product 999 from products service: rpc error: code = NotFound desc = product not found\nproduct not found for the given id")
	assert.Empty(t, updated)
}

func TestUpdateOrder_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.expectFindByID(
		func(_ context.Context, id uuid.UUID) (order.Order, error) {
			if id == testO.ID {
				return testO, nil
			} else {
				return order.Order{}, errors.New("wrong id passed in fake")
			}
		})
	mock.expectUpdate(
		func(_ context.Context, o order.Order) (order.Order, error) {
			return o, errors.New("update order: database unavailable")
		})

	updated, err := svc.UpdateOrder(context.Background(), testO.ID, FirstProductID, 10, order.CANCELED)

	assert.Error(t, err)
	assert.EqualError(t, err, "updating order: update order: database unavailable")
	assert.Empty(t, updated)
}

func TestDeleteOrder_WhenOrderNotExists_DoesNotFail(t *testing.T) {
	svc, mock, _ := setup(t)
	toDelete, _ := uuid.NewV7()
	mock.expectDelete(
		func(_ context.Context, id uuid.UUID) error {
			if id == toDelete {
				return nil
			} else {
				return errors.New("wrong id passed in fake")
			}
		})

	err := svc.DeleteOrder(context.Background(), toDelete)

	assert.NoError(t, err)
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
}

var (
	FirstProductID  = "f47ac10b-58cc-4372-a567-0e02b2c3d001"
	SecondProductID = "f47ac10b-58cc-4372-a567-0e02b2c3d002"

	FirstOrderID  = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d011")
	SecondOrderID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d012")

	testO = order.Order{
		ID:        FirstOrderID,
		ProductID: FirstProductID,
		Quantity:  2,
		Status:    order.CREATED,
		CreatedAt: time.Now(),
	}
)
