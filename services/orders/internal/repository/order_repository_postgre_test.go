package repository

import (
	"commerce-platform/services/orders/internal/order"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

func setup(t *testing.T) (pgxmock.PgxPoolIface, *PostgreOrderRepository) {
	t.Helper()
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)

	t.Cleanup(func() {
		mock.Close()
	})

	return mock, NewPostgreOrderRepository(mock)
}

func TestFindAll_WhenOrdersExist_ReturnsAllOrders(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`SELECT order_id, product_id, quantity, status, created_at FROM orders`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at",
			}).
				AddRow(FirstOrderID, FirstProductID, 2, order.CREATED, time.Now()).
				AddRow(SecondOrderID, SecondProductID, 1, order.PAID, time.Now()),
		)

	result, err := repo.FindAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindAll_WhenDbError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`SELECT order_id, product_id, quantity, status, created_at FROM orders`).
		WillReturnError(errors.New("database unavailable"))

	result, err := repo.FindAll(context.Background())

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindAll_WhenScanError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`SELECT order_id, product_id, quantity, status, created_at FROM orders`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at",
			}).
				AddRow("invalid-uuid", "invalid-uuid", 2, order.CREATED, time.Now()),
		)

	result, err := repo.FindAll(context.Background())

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindAll_WhenRowsError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`SELECT order_id, product_id, quantity, status, created_at FROM orders`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at",
			}).
				AddRow(FirstOrderID, FirstProductID, 2, order.CREATED, time.Now()).
				RowError(1, errors.New("read failed")),
		)

	result, err := repo.FindAll(context.Background())

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByID_WhenOrdersExist_ReturnsOrder(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at
			FROM orders
			WHERE order_id = $1`).
		WithArgs(FirstOrderID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at",
			}).
				AddRow(testO.ID, testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt),
		)

	result, err := repo.FindByID(context.Background(), FirstOrderID)

	assert.NoError(t, err)
	assert.Equal(t, testO, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByID_WhenDbError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at
			FROM orders
			WHERE order_id = $1`).
		WithArgs(FirstOrderID).
		WillReturnError(errors.New("database unavailable"))

	result, err := repo.FindByID(context.Background(), FirstOrderID)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByID_WhenScanError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at
			FROM orders
			WHERE order_id = $1`).
		WithArgs(FirstOrderID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at",
			}).
				AddRow("invalid-uuid", "invalid-uuid", 2, order.CREATED, time.Now()),
		)

	result, err := repo.FindByID(context.Background(), FirstOrderID)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByID_WhenRowsError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at
			FROM orders
			WHERE order_id = $1`).
		WithArgs(FirstOrderID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at",
			}).
				RowError(1, pgx.ErrNoRows),
		)

	result, err := repo.FindByID(context.Background(), FirstOrderID)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

var (
	testO = order.Order{
		ID:        FirstOrderID,
		ProductID: FirstProductID,
		Quantity:  2,
		Status:    order.CREATED,
		CreatedAt: time.Now(),
	}
)
