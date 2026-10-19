package http

import (
	"bytes"
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/services/orders/internal/repository"
	"commerce-platform/services/orders/internal/service"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

// happy path and cutting cross concerns tests

func setupOrderHandlerCompositionTest(t *testing.T) (*httptest.Server, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)

	t.Cleanup(func() {
		mock.Close()
	})
	repo := repository.NewPostgreOrderRepository(mock, repository.Timeouts{
		Read:  100 * time.Millisecond,
		List:  100 * time.Second,
		Write: 100 * time.Millisecond,
	})
	client := &mockProductsClient{
		productIDs: map[string]bool{
			FirstProductID:  true,
			SecondProductID: true,
		},
	}
	svc := service.NewOrderService(repo, client)
	handler := NewOrderHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	return srv, mock
}

func TestGetOrdersComposition_WhenOrdersExist_Returns200(t *testing.T) {
	srv, mock := setupOrderHandlerCompositionTest(t)
	mock.ExpectQuery(`SELECT order_id, product_id, quantity, status, created_at, version FROM orders`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(FirstOrderID, FirstProductID, 2, order.CREATED, time.Now(), 1).
				AddRow(SecondOrderID, SecondProductID, 1, order.PAID, time.Now(), 1),
		)

	res, err := http.Get(srv.URL + "/orders")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var resOrders []map[string]any
	err = json.Unmarshal(body, &resOrders)
	assert.NoError(t, err)

	expectedOrders := []map[string]any{
		{
			"id":         FirstOrderID.String(),
			"product_id": FirstProductID,
			"quantity":   float64(2),
			"status":     "CREATED",
			"version":    float64(1),
		},
		{
			"id":         SecondOrderID.String(),
			"product_id": SecondProductID,
			"quantity":   float64(1),
			"status":     "PAID",
			"version":    float64(1),
		},
	}

	// remove createdAt field from the comparison
	for _, order := range resOrders {
		delete(order, "created_at")
	}

	assert.ElementsMatch(t, expectedOrders, resOrders)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrderComposition_WhenOrderNotExists_Returns404(t *testing.T) {
	srv, mock := setupOrderHandlerCompositionTest(t)
	id, _ := uuid.NewV7()
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at, version
			FROM orders
			WHERE order_id = $1`).
		WithArgs(id).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}))

	res, err := http.Get(srv.URL + "/orders/" + id.String())

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "ORDER_NOT_FOUND",
			"message": "Order with id [`+id.String()+`] was not found"
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrderComposition_WhenQueryExceedsTimeout_Returns503(t *testing.T) {
	srv, mock := setupOrderHandlerCompositionTest(t)
	id, _ := uuid.NewV7()
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at, version
			FROM orders
			WHERE order_id = $1`).
		WithArgs(id).
		WillDelayFor(120 * time.Millisecond).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}))

	res, err := http.Get(srv.URL + "/orders/" + id.String())

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "REQUEST_TIMED_OUT",
			"message": "request timed out"
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrderComposition_WhenRequestValid_CreatesOrder(t *testing.T) {
	srv, mock := setupOrderHandlerCompositionTest(t)
	mock.ExpectQuery(`
			INSERT INTO orders (order_id, product_id, quantity, status)
			VALUES ($1, $2, $3, $4)
			RETURNING order_id, product_id, quantity, status, created_at, version;`).
		WithArgs(pgxmock.OfType[uuid.UUID](), FirstProductID, 10, order.CREATED).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(uuid.New(), FirstProductID, 10, order.CREATED, time.Now(), 1),
		)

	res, err := http.Post(
		srv.URL+"/orders",
		"application/json",
		bytes.NewBufferString(`{
			"product_id": " `+FirstProductID+` ",
			"quantity": 10
		}`),
	)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	var created order.Order
	err = json.Unmarshal(body, &created)
	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID) // a UUID was assigned
	assert.Equal(t, FirstProductID, created.ProductID)
	assert.Equal(t, 10, created.Quantity)
	assert.Equal(t, order.CREATED, created.Status)
	assert.WithinDuration(t, time.Now(), created.CreatedAt, time.Second)
	assert.Equal(t, 1, created.Version)
	assert.Equal(t, "/orders/"+created.ID.String(), res.Header.Get("Location"))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrderComposition_WhenProductNotExists_Returns409(t *testing.T) {
	srv, mock := setupOrderHandlerCompositionTest(t)

	res, err := http.Post(
		srv.URL+"/orders",
		"application/json",
		bytes.NewBufferString(`{
			"product_id": "`+validUUID+`",
			"quantity": 1
		}`),
	)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusConflict, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "PRODUCT_NOT_FOUND",
			"message": "product not found for the given id"
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrderComposition_WhenRequestValid_UpdatesOrder(t *testing.T) {
	srv, mock := setupOrderHandlerCompositionTest(t)
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at, version
			FROM orders
			WHERE order_id = $1`).
		WithArgs(testO.ID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(testO.ID, testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt, testO.Version),
		)
	mock.ExpectQuery(`
			UPDATE orders
			SET product_id = $1, quantity = $2, status = $3, version = version + 1
			WHERE order_id = $4
			  AND version = $5
			RETURNING order_id, product_id, quantity, status, created_at, version;`).
		WithArgs(FirstProductID, 11, order.PAID, testO.ID, testO.Version).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(testO.ID, FirstProductID, 11, order.PAID, testO.CreatedAt, testO.Version+1),
		)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+FirstOrderID.String(),
		bytes.NewBufferString(`{
			"product_id": " `+FirstProductID+`",
			"quantity": 11,
			"status": "paid",
			"version": 1
		}`),
	)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	var updated order.Order
	err = json.Unmarshal(body, &updated)
	assert.NoError(t, err)
	assert.Equal(t, testO.ID, updated.ID)
	assert.Equal(t, FirstProductID, updated.ProductID)
	assert.Equal(t, 11, updated.Quantity)
	assert.Equal(t, order.PAID, updated.Status)
	assert.True(t, updated.CreatedAt.Equal(testO.CreatedAt))
	assert.Equal(t, testO.Version+1, updated.Version)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrderComposition_WhenProductNotExists_Returns409(t *testing.T) {
	srv, mock := setupOrderHandlerCompositionTest(t)
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at, version
			FROM orders
			WHERE order_id = $1`).
		WithArgs(testO.ID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(testO.ID, testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt, testO.Version),
		)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+testO.ID.String(),
		bytes.NewBufferString(`{
			"product_id": "`+validUUID+`",
			"quantity": 2,
			"status": "PAID",
			"version": 1
		}`),
	)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusConflict, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "PRODUCT_NOT_FOUND",
			"message": "product not found for the given id"
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrderComposition_WhenOrderNotExists_Returns404(t *testing.T) {
	srv, mock := setupOrderHandlerCompositionTest(t)
	id, _ := uuid.NewV7()
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at, version
			FROM orders
			WHERE order_id = $1`).
		WithArgs(id).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}))

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+id.String(),
		bytes.NewBufferString(`{
			"product_id": "`+FirstProductID+`",
			"quantity": 1,
			"status": "PAiD",
			"version": 1
		}`),
	)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "ORDER_NOT_FOUND",
			"message": "Order with id [`+id.String()+`] was not found"
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrderComposition_WhenUpdateConflict_Returns409(t *testing.T) {
	srv, mock := setupOrderHandlerCompositionTest(t)
	mock.ExpectQuery(`
			SELECT order_id, product_id, quantity, status, created_at, version
			FROM orders
			WHERE order_id = $1`).
		WithArgs(testO.ID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(testO.ID, testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt, testO.Version),
		)
	mock.ExpectQuery(`
			UPDATE orders
			SET product_id = $1, quantity = $2, status = $3, version = version + 1
			WHERE order_id = $4
			  AND version = $5
			RETURNING order_id, product_id, quantity, status, created_at, version;`).
		WithArgs(FirstProductID, 11, order.PAID, testO.ID, testO.Version).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at", "version",
			}).
				AddRow(testO.ID, FirstProductID, 11, order.PAID, testO.CreatedAt, testO.Version+1).
				RowError(0, pgx.ErrNoRows),
		)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+FirstOrderID.String(),
		bytes.NewBufferString(`{
			"product_id": " `+FirstProductID+`",
			"quantity": 11,
			"status": "paid",
			"version": 1
		}`),
	)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusConflict, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "ORDER_UPDATE_CONFLICT",
			"message": "order was updated by another process"
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteOrderComposition_WhenOrderExists_DeletesOrder(t *testing.T) {
	srv, mock := setupOrderHandlerCompositionTest(t)
	mock.ExpectExec(`
			DELETE FROM orders
			WHERE order_id = $1`).
		WithArgs(SecondOrderID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	req, err := http.NewRequest(
		http.MethodDelete,
		srv.URL+"/orders/"+SecondOrderID.String(),
		nil,
	)
	assert.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, res.StatusCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}
