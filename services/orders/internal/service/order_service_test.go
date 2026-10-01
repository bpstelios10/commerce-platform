package service

import (
	"commerce-platform/services/orders/internal/grpc"
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/services/orders/internal/repository"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v5"
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

func setup(t *testing.T) (*OrderService, pgxmock.PgxPoolIface, *mockProductsClient) {
	t.Helper()
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)

	t.Cleanup(func() {
		mock.Close()
	})
	repo := repository.NewPostgreOrderRepository(mock)
	client := &mockProductsClient{
		productIDs: map[string]bool{
			FirstProductID:  true,
			SecondProductID: true,
		},
	}
	svc := NewOrderService(repo, client)

	return svc, mock, client
}

func TestGetOrders_WhenOrdersExist_ReturnsOrders(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.ExpectQuery(`SELECT order_id, product_id, quantity, status, created_at FROM orders`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at",
			}).
				AddRow(FirstOrderID, FirstProductID, 2, order.CREATED, time.Now()).
				AddRow(SecondOrderID, SecondProductID, 1, order.PAID, time.Now()),
		)

	orders, err := svc.GetOrders(context.Background())

	assert.NoError(t, err)
	assert.Len(t, orders, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrders_WhenOtherError_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.ExpectQuery(`SELECT order_id, product_id, quantity, status, created_at FROM orders`).
		WillReturnError(errors.New("database unavailable"))

	orders, err := svc.GetOrders(context.Background())

	assert.Error(t, err)
	assert.EqualError(t, err, "get orders: query orders: database unavailable")
	assert.Empty(t, orders)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrderByID_WhenOrderExists_ReturnsOrder(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at
			FROM orders
			WHERE order_id = $1`).
		WithArgs(testO.ID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at",
			}).
				AddRow(testO.ID, testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt),
		)

	o, err := svc.GetOrderByID(context.Background(), testO.ID)

	assert.NoError(t, err)
	assert.Equal(t, testO, o)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrderByID_WhenOrderNotExists_ReturnsNotFound(t *testing.T) {
	svc, mock, _ := setup(t)
	id, _ := uuid.NewV7()
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at
			FROM orders
			WHERE order_id = $1`).
		WithArgs(id).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at",
			}))

	o, err := svc.GetOrderByID(context.Background(), id)

	assert.Error(t, err)
	assert.EqualError(t, err, "Order with id ["+id.String()+"] was not found")
	var notFoundErr *ErrOrderNotFound
	assert.ErrorAs(t, err, &notFoundErr)
	assert.Equal(t, id, notFoundErr.OrderID)
	assert.Empty(t, o)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrderByID_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	id, _ := uuid.NewV7()
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at
			FROM orders
			WHERE order_id = $1`).
		WithArgs(id).
		WillReturnError(errors.New("database unavailable"))

	o, err := svc.GetOrderByID(context.Background(), id)

	assert.Error(t, err)
	assert.EqualError(t, err, "get order by id: find order by id: database unavailable")
	assert.Empty(t, o)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrder_WhenProductExists_CreatesOrder(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.ExpectQuery(`
			INSERT INTO orders (order_id, product_id, quantity, status)
			VALUES ($1, $2, $3, $4)
			RETURNING order_id, product_id, quantity, status, created_at;`).
		WithArgs(pgxmock.OfType[uuid.UUID](), FirstProductID, 10, order.CREATED).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at",
			}).
				AddRow(uuid.New(), FirstProductID, 10, order.CREATED, time.Now()),
		)

	o, err := svc.CreateOrder(context.Background(), FirstProductID, 10)

	assert.NoError(t, err)
	assert.NotNil(t, o.ID)
	assert.Equal(t, FirstProductID, o.ProductID)
	assert.Equal(t, 10, o.Quantity)
	assert.WithinDuration(t, time.Now(), o.CreatedAt, time.Second)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrder_WhenProductNotExists_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)

	o, err := svc.CreateOrder(context.Background(), "999", 10)

	assert.ErrorIs(t, err, ErrProductNotFound)
	assert.Empty(t, o)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrder_WhenProductValidationFails_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)

	o, err := svc.CreateOrder(context.Background(), "error", 10)

	assert.Error(t, err)
	assert.EqualError(t, err, "create order: get product error from products service: rpc error: code = Internal desc = unexpected error")
	assert.Empty(t, o)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrder_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.ExpectQuery(`
			INSERT INTO orders (order_id, product_id, quantity, status)
			VALUES ($1, $2, $3, $4)
			RETURNING order_id, product_id, quantity, status, created_at;`).
		WithArgs(pgxmock.OfType[uuid.UUID](), FirstProductID, 10, order.CREATED).
		WillReturnError(errors.New("database unavailable"))

	o, err := svc.CreateOrder(context.Background(), FirstProductID, 10)

	assert.Error(t, err)
	assert.EqualError(t, err, "create order: save order: database unavailable")
	assert.Empty(t, o)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrder_WhenOrderExists_UpdatesOrder(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at
			FROM orders
			WHERE order_id = $1`).
		WithArgs(testO.ID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at",
			}).
				AddRow(testO.ID, testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt),
		)
	mock.ExpectQuery(`
			UPDATE orders
			SET product_id = $1, quantity = $2, status = $3
			WHERE order_id = $4
			RETURNING order_id, product_id, quantity, status, created_at;`).
		WithArgs(FirstProductID, 11, order.PAID, testO.ID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at",
			}).
				AddRow(testO.ID, FirstProductID, 11, order.PAID, testO.CreatedAt),
		)

	updated, err := svc.UpdateOrder(context.Background(), testO.ID, FirstProductID, 11, order.PAID)

	assert.NoError(t, err)
	assert.Equal(t, testO.ID, updated.ID)
	assert.Equal(t, FirstProductID, updated.ProductID)
	assert.Equal(t, 11, updated.Quantity)
	assert.Equal(t, order.PAID, updated.Status)
	assert.True(t, testO.CreatedAt.Equal(updated.CreatedAt))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrder_WhenOrderNotExists_CreatesOrder(t *testing.T) {
	svc, mock, _ := setup(t)
	id := uuid.New()
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at
			FROM orders
			WHERE order_id = $1`).
		WithArgs(id).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at",
			}))

	updated, err := svc.UpdateOrder(context.Background(), id, "1", 10, order.CANCELED)

	assert.Error(t, err)
	assert.EqualError(t, err, "updating order: Order with id ["+id.String()+"] was not found")
	assert.Empty(t, updated)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrder_WhenProductNotExists_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at
			FROM orders
			WHERE order_id = $1`).
		WithArgs(FirstOrderID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at",
			}).
				AddRow(FirstOrderID, FirstProductID, 2, order.CREATED, time.Now()),
		)

	updated, err := svc.UpdateOrder(context.Background(), FirstOrderID, "999", 11, order.PAID)

	assert.Error(t, err)
	assert.EqualError(t, err, "updating order: get product 999 from products service: rpc error: code = NotFound desc = product not found\nproduct not found for the given id")
	assert.Empty(t, updated)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrder_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at
			FROM orders
			WHERE order_id = $1`).
		WithArgs(testO.ID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at",
			}).
				AddRow(testO.ID, testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt),
		)
	mock.ExpectQuery(`
			UPDATE orders
			SET product_id = $1, quantity = $2, status = $3
			WHERE order_id = $4
			RETURNING order_id, product_id, quantity, status, created_at;`).
		WithArgs(FirstProductID, 10, order.CANCELED, testO.ID).
		WillReturnError(errors.New("database unavailable"))

	updated, err := svc.UpdateOrder(context.Background(), testO.ID, FirstProductID, 10, order.CANCELED)

	assert.Error(t, err)
	assert.EqualError(t, err, "updating order: update order: database unavailable")
	assert.Empty(t, updated)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteOrder_WhenOrderNotExists_DoesNotFail(t *testing.T) {
	svc, mock, _ := setup(t)
	id, _ := uuid.NewV7()
	mock.ExpectExec(`
			DELETE FROM orders
			WHERE order_id = $1`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := svc.DeleteOrder(context.Background(), id)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteOrder_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock, _ := setup(t)
	id, _ := uuid.NewV7()
	mock.ExpectExec(`
			DELETE FROM orders
			WHERE order_id = $1`).
		WithArgs(id).
		WillReturnError(errors.New("database unavailable"))

	err := svc.DeleteOrder(context.Background(), id)

	assert.Error(t, err)
	assert.EqualError(t, err, "deleting order: delete order: database unavailable")
	assert.NoError(t, mock.ExpectationsWereMet())
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
