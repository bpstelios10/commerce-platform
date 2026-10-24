package repository

import (
	"commerce-platform/services/orders/internal/order"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

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

	return mock, NewPostgreOrderRepository(mock, Timeouts{
		Read:  100 * time.Millisecond,
		List:  100 * time.Millisecond,
		Write: 100 * time.Millisecond,
	})
}

func TestList_WhenOrdersExist_ReturnsAllOrders(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`SELECT order_id, product_id, quantity, status, created_at, version FROM orders`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(FirstOrderID, FirstProductID, 2, order.CREATED, time.Now(), 1).
				AddRow(SecondOrderID, SecondProductID, 1, order.PAID, time.Now(), 1),
		)

	result, err := repo.List(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestList_WhenDbError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`SELECT order_id, product_id, quantity, status, created_at, version FROM orders`).
		WillReturnError(errors.New("database unavailable"))

	result, err := repo.List(context.Background())

	assert.Error(t, err)
	assert.EqualError(t, err, "query orders: database unavailable")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestList_WhenScanError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`SELECT order_id, product_id, quantity, status, created_at, version FROM orders`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow("invalid-uuid", "invalid-uuid", 2, order.CREATED, time.Now(), 1),
		)

	result, err := repo.List(context.Background())

	assert.Error(t, err)
	assert.EqualError(t, err, "scan order: scanning value error for column 'order_id': Scan: invalid UUID length: 12")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestList_WhenRowsError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`SELECT order_id, product_id, quantity, status, created_at, version FROM orders`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(FirstOrderID, FirstProductID, 2, order.CREATED, time.Now(), 1).
				RowError(1, errors.New("read failed")),
		)

	result, err := repo.List(context.Background())

	assert.Error(t, err)
	assert.EqualError(t, err, "iterate orders: read failed")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestList_WhenQueryExceedsTimeout_ReturnsDeadlineExceeded(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`SELECT order_id, product_id, quantity, status, created_at, version FROM orders`).
		WillDelayFor(120 * time.Millisecond).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(FirstOrderID, FirstProductID, 2, order.CREATED, time.Now(), 1).
				AddRow(SecondOrderID, SecondProductID, 1, order.PAID, time.Now(), 1),
		)

	result, err := repo.List(context.Background())

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.EqualError(t, err, "query orders: context deadline exceeded")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_WhenOrdersExist_ReturnsOrder(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at, version
			FROM orders
			WHERE order_id = $1`).
		WithArgs(FirstOrderID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(testO.ID, testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt, testO.Version),
		)

	result, err := repo.GetByID(context.Background(), FirstOrderID)

	assert.NoError(t, err)
	assert.Equal(t, testO, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_WhenErrorNotExists_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at, version
			FROM orders
			WHERE order_id = $1`).
		WithArgs(FirstOrderID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(FirstOrderID, FirstProductID, 2, order.CREATED, time.Now(), 1).
				RowError(0, pgx.ErrNoRows),
		)

	result, err := repo.GetByID(context.Background(), FirstOrderID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
	assert.EqualError(t, err, "find order by id: not found")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_WhenDbError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at, version
			FROM orders
			WHERE order_id = $1`).
		WithArgs(FirstOrderID).
		WillReturnError(errors.New("database unavailable"))

	result, err := repo.GetByID(context.Background(), FirstOrderID)

	assert.Error(t, err)
	assert.EqualError(t, err, "find order by id: database unavailable")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_WhenScanError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at, version
			FROM orders
			WHERE order_id = $1`).
		WithArgs(FirstOrderID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow("invalid-uuid", "invalid-uuid", 2, order.CREATED, time.Now(), 1),
		)

	result, err := repo.GetByID(context.Background(), FirstOrderID)

	assert.Error(t, err)
	assert.EqualError(t, err, "find order by id: scanning value error for column 'order_id': Scan: invalid UUID length: 12")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_WhenQueryExceedsTimeout_ReturnsDeadlineExceeded(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
            SELECT order_id, product_id, quantity, status, created_at, version
            FROM orders
            WHERE order_id = $1`).
		WithArgs(FirstOrderID).
		WillDelayFor(120 * time.Millisecond).
		WillReturnRows(pgxmock.NewRows([]string{
			"order_id", "product_id", "quantity", "status", "created_at", "version",
		}).AddRow(testO.ID, testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt, testO.Version))

	result, err := repo.GetByID(context.Background(), FirstOrderID)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.EqualError(t, err, "find order by id: context deadline exceeded")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_WhenNoDbIssues_ReturnsNewOrder(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			INSERT INTO orders (order_id, product_id, quantity, status)
			VALUES ($1, $2, $3, $4)
			RETURNING order_id, product_id, quantity, status, created_at, version;`).
		WithArgs(testO.ID, testO.ProductID, testO.Quantity, testO.Status).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(testO.ID, testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt, testO.Version),
		)

	result, err := repo.Create(context.Background(), testO)

	assert.NoError(t, err)
	assert.Equal(t, testO, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_WhenDbError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			INSERT INTO orders (order_id, product_id, quantity, status)
			VALUES ($1, $2, $3, $4)
			RETURNING order_id, product_id, quantity, status, created_at, version;`).
		WithArgs(testO.ID, testO.ProductID, testO.Quantity, testO.Status).
		WillReturnError(errors.New("database unavailable"))

	result, err := repo.Create(context.Background(), testO)

	assert.Error(t, err)
	assert.EqualError(t, err, "save order: database unavailable")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_WhenScanError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			INSERT INTO orders (order_id, product_id, quantity, status)
			VALUES ($1, $2, $3, $4)
			RETURNING order_id, product_id, quantity, status, created_at, version;`).
		WithArgs(testO.ID, testO.ProductID, testO.Quantity, testO.Status).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow("invalid-uuid", testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt, testO.Version),
		)

	result, err := repo.Create(context.Background(), testO)

	assert.Error(t, err)
	assert.EqualError(t, err, "save order: scanning value error for column 'order_id': Scan: invalid UUID length: 12")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_WhenRowsError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			INSERT INTO orders (order_id, product_id, quantity, status)
			VALUES ($1, $2, $3, $4)
			RETURNING order_id, product_id, quantity, status, created_at, version;`).
		WithArgs(testO.ID, testO.ProductID, testO.Quantity, testO.Status).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(FirstOrderID, FirstProductID, 2, order.CREATED, time.Now(), 1).
				RowError(0, errors.New("read failed")),
		)

	result, err := repo.Create(context.Background(), testO)

	assert.Error(t, err)
	assert.EqualError(t, err, "save order: read failed\nread failed")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_WhenQueryExceedsTimeout_ReturnsDeadlineExceeded(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			INSERT INTO orders (order_id, product_id, quantity, status)
			VALUES ($1, $2, $3, $4)
			RETURNING order_id, product_id, quantity, status, created_at, version;`).
		WithArgs(testO.ID, testO.ProductID, testO.Quantity, testO.Status).
		WillDelayFor(120 * time.Millisecond).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(testO.ID, testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt, testO.Version),
		)

	result, err := repo.Create(context.Background(), testO)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.EqualError(t, err, "save order: context deadline exceeded")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_WhenOrderExists_ReturnsUpdatedOrder(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			UPDATE orders
			SET product_id = $1, quantity = $2, status = $3, version = version + 1
			WHERE order_id = $4
			  AND version = $5
			RETURNING order_id, product_id, quantity, status, created_at, version;`).
		WithArgs(testO.ProductID, testO.Quantity, testO.Status, testO.ID, testO.Version).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(testO.ID, testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt, testO.Version+1),
		)

	result, err := repo.Update(context.Background(), testO)

	assert.NoError(t, err)
	testO.Version = testO.Version + 1
	assert.Equal(t, testO, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_WhenUpdateConflict_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			UPDATE orders
			SET product_id = $1, quantity = $2, status = $3, version = version + 1
			WHERE order_id = $4
			  AND version = $5
			RETURNING order_id, product_id, quantity, status, created_at, version;`).
		WithArgs(testO.ProductID, testO.Quantity, testO.Status, testO.ID, testO.Version).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(FirstOrderID, FirstProductID, 2, order.CREATED, time.Now(), 2).
				RowError(0, pgx.ErrNoRows),
		)

	result, err := repo.Update(context.Background(), testO)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrVersionConflict)
	assert.EqualError(t, err, "update order: order was updated by another transaction")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_WhenDbError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			UPDATE orders
			SET product_id = $1, quantity = $2, status = $3, version = version + 1
			WHERE order_id = $4
			  AND version = $5
			RETURNING order_id, product_id, quantity, status, created_at, version;`).
		WithArgs(testO.ProductID, testO.Quantity, testO.Status, testO.ID, testO.Version).
		WillReturnError(errors.New("database unavailable"))

	result, err := repo.Update(context.Background(), testO)

	assert.Error(t, err)
	assert.EqualError(t, err, "update order: database unavailable")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_WhenScanError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			UPDATE orders
			SET product_id = $1, quantity = $2, status = $3, version = version + 1
			WHERE order_id = $4
			  AND version = $5
			RETURNING order_id, product_id, quantity, status, created_at, version;`).
		WithArgs(testO.ProductID, testO.Quantity, testO.Status, testO.ID, testO.Version).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow("invalid-uuid", testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt, testO.Version+1),
		)

	result, err := repo.Update(context.Background(), testO)

	assert.Error(t, err)
	assert.EqualError(t, err, "update order: scanning value error for column 'order_id': Scan: invalid UUID length: 12")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_WhenQueryExceedsTimeout_ReturnsDeadlineExceeded(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			UPDATE orders
			SET product_id = $1, quantity = $2, status = $3, version = version + 1
			WHERE order_id = $4
			  AND version = $5
			RETURNING order_id, product_id, quantity, status, created_at, version;`).
		WithArgs(testO.ProductID, testO.Quantity, testO.Status, testO.ID, testO.Version).
		WillDelayFor(120 * time.Millisecond).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(testO.ID, testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt, testO.Version+1),
		)

	result, err := repo.Update(context.Background(), testO)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.EqualError(t, err, "update order: context deadline exceeded")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_WhenNoDbIssues_Returns(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectExec(`
			DELETE FROM orders
			WHERE order_id = $1`).
		WithArgs(SecondOrderID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := repo.Delete(context.Background(), SecondOrderID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_WhenDbError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectExec(`
			DELETE FROM orders
			WHERE order_id = $1`).
		WithArgs(SecondOrderID).
		WillReturnError(errors.New("database unavailable"))

	err := repo.Delete(context.Background(), SecondOrderID)

	assert.Error(t, err)
	assert.EqualError(t, err, "delete order: database unavailable")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_WhenQueryExceedsTimeout_ReturnsDeadlineExceeded(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectExec(`
			DELETE FROM orders
			WHERE order_id = $1`).
		WithArgs(SecondOrderID).
		WillDelayFor(120 * time.Millisecond).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := repo.Delete(context.Background(), SecondOrderID)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.EqualError(t, err, "delete order: context deadline exceeded")
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
		Version:   1,
	}
)
