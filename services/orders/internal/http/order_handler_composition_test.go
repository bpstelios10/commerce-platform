package http

import (
	"bytes"
	"commerce-platform/services/orders/internal/order"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

// happy path and cutting cross concerns tests

func TestGetOrdersComposition_WhenOrdersExist_Returns200(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
	mock.ExpectQuery(`SELECT order_id, product_id, quantity, status, created_at FROM orders`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"order_id", "product_id", "quantity", "status", "created_at",
			}).
				AddRow(FirstOrderID, FirstProductID, 2, order.CREATED, time.Now()).
				AddRow(SecondOrderID, SecondProductID, 1, order.PAID, time.Now()),
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
		},
		{
			"id":         SecondOrderID.String(),
			"product_id": SecondProductID,
			"quantity":   float64(1),
			"status":     "PAID",
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
	srv, mock := setupOrderHandlerTest(t)
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

func TestCreateOrderComposition_WhenRequestValid_CreatesOrder(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
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
	assert.Equal(t, "/orders/"+created.ID.String(), res.Header.Get("Location"))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrderComposition_WhenProductNotExists_Returns409(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)

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
	srv, mock := setupOrderHandlerTest(t)
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

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+FirstOrderID.String(),
		bytes.NewBufferString(`{
			"product_id": " `+FirstProductID+`",
			"quantity": 11,
			"status": "paid"
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
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrderComposition_WhenProductNotExists_Returns409(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
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

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+testO.ID.String(),
		bytes.NewBufferString(`{
			"product_id": "`+validUUID+`",
			"quantity": 2,
			"status": "PAID"
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
	srv, mock := setupOrderHandlerTest(t)
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

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+id.String(),
		bytes.NewBufferString(`{
			"product_id": "`+FirstProductID+`",
			"quantity": 1,
			"status": "PAiD"
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

func TestDeleteOrderComposition_WhenOrderExists_DeletesOrder(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
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

	assert.Equal(t, http.StatusNoContent, res.StatusCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}
