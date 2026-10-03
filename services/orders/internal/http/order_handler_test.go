package http

import (
	"bytes"
	"commerce-platform/services/orders/internal/grpc"
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/services/orders/internal/repository"
	"commerce-platform/services/orders/internal/service"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

type mockProductsClient struct {
	productIDs map[string]bool
}

func (m *mockProductsClient) GetProductByID(_ context.Context, id string) (*grpc.GetProductByIDResponse, error) {
	if m.productIDs[id] {
		return &grpc.GetProductByIDResponse{Id: id}, nil
	}
	return nil, service.ErrProductNotFound
}

// To be used as BeforeEach
func setupOrderHandlerTest(t *testing.T) (*httptest.Server, pgxmock.PgxPoolIface) {
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
	svc := service.NewOrderService(repo, client)
	handler := NewOrderHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	return srv, mock
}

// To be used as BeforeEach
func setupOrderMuxHandlerTest(t *testing.T) (*chi.Mux, pgxmock.PgxPoolIface) {
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
	svc := service.NewOrderService(repo, client)
	handler := NewOrderHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	return r, mock
}

func TestGetOrders_WhenOrdersExist_Returns200(t *testing.T) {
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

func TestGetOrder_WhenDbErrorHappens_Returns500(t *testing.T) {
	r, mock := setupOrderMuxHandlerTest(t)
	mock.ExpectQuery(`SELECT order_id, product_id, quantity, status, created_at FROM orders`).
		WillReturnError(errors.New("database unavailable"))

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders",
		nil,
	)
	req = req.WithContext(context.WithValue(req.Context(), "errorEnabler", "unexpected error"))
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INTERNAL_SERVER_ERROR",
			"message": "internal server error"
		}`,
		res.Body.String(),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrder_WhenOrderExists_Returns200(t *testing.T) {
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

	res, err := http.Get(srv.URL + "/orders/" + testO.ID.String())

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	var existing order.Order
	err = json.Unmarshal(body, &existing)
	assert.NoError(t, err)
	assert.Equal(t, testO.ID, existing.ID)
	assert.Equal(t, testO.ProductID, existing.ProductID)
	assert.Equal(t, testO.Quantity, existing.Quantity)
	assert.Equal(t, testO.Status, existing.Status)
	assert.WithinDuration(t, testO.CreatedAt, existing.CreatedAt, time.Millisecond)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrder_WhenOrderNotExists_Returns404(t *testing.T) {
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

func TestGetOrder_WhenBadUUID_Returns400(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)

	res, err := http.Get(srv.URL + "/orders/1234")
	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INVALID_UUID",
			"message": "invalid UUID"
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrder_WhenRequestValid_CreatesOrder(t *testing.T) {
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
			"product_id": "`+FirstProductID+`",
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

func TestCreateOrder_WhenProductNotExists_Returns409(t *testing.T) {
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

func TestCreateOrder_WhenBadRequestBody_Returns400(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)

	res, err := http.Post(
		srv.URL+"/orders",
		"application/json",
		bytes.NewBufferString(`{
			"error-to-cause": "extra comma, so invalid json",
		}`),
	)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INVALID_ORDER",
			"message": "invalid order"
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrder_WhenRequestInvalid_Returns400(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)

	res, err := http.Post(
		srv.URL+"/orders",
		"application/json",
		bytes.NewBufferString(`{
			"product_id": "",
			"quantity": 0
		}`),
	)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "VALIDATION_ERROR",
			"message": "product-id cannot be blank.; quantity must be > 0."
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrder_WhenRequestValid_UpdatesOrder(t *testing.T) {
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
			"product_id": "`+FirstProductID+`",
			"quantity": 11,
			"status": "PAID"
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

func TestUpdateOrder_WhenRequestValidWithLowercaseStatus_UpdatesOrder(t *testing.T) {
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
			"product_id": "`+FirstProductID+`",
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

func TestUpdateOrder_WhenProductNotExists_Returns409(t *testing.T) {
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

func TestUpdateOrder_WhenBadRequestBody_Returns400(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+FirstOrderID.String(),
		bytes.NewBufferString(`{
			"error-to-cause": "extra comma, so invalid json",
		}`),
	)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INVALID_ORDER",
			"message": "invalid order"
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrder_WhenRequestInvalid_Returns400(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+FirstOrderID.String(),
		bytes.NewBufferString(`{
			"product_id": "",
			"quantity": 0,
			"status": "PIAD"
		}`),
	)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "VALIDATION_ERROR",
			"message": "product-id cannot be blank.; quantity must be > 0.; status is not valid."
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrder_WhenBadUUID_Returns400(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/1234",
		bytes.NewBufferString(`{
			"product_id": "`+FirstProductID+`",
			"quantity": 1,
			"status": "PAID"
		}`),
	)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INVALID_UUID",
			"message": "invalid UUID"
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrder_WhenOrderNotExists_Returns404(t *testing.T) {
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

func TestDeleteOrder_WhenOrderExists_DeletesOrder(t *testing.T) {
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

func TestDeleteOrder_WhenBadUUID_Returns400(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodDelete,
		srv.URL+"/orders/1234",
		nil,
	)
	assert.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INVALID_UUID",
			"message": "invalid UUID"
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteOrder_WhenDbErrorHappens_Returns500(t *testing.T) {
	r, mock := setupOrderMuxHandlerTest(t)
	mock.ExpectExec(`
			DELETE FROM orders
			WHERE order_id = $1`).
		WithArgs(SecondOrderID).
		WillReturnError(errors.New("database unavailable"))

	req := httptest.NewRequest(
		http.MethodDelete,
		"/orders/"+SecondOrderID.String(),
		nil,
	)
	req = req.WithContext(context.WithValue(req.Context(), "errorEnabler", "unexpected error"))
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	body, _ := io.ReadAll(res.Body)
	assert.JSONEq(
		t,
		`{
			"code": "INTERNAL_SERVER_ERROR",
			"message": "internal server error"
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

var (
	FirstProductID  = "f47ac10b-58cc-4372-a567-0e02b2c3d001"
	SecondProductID = "f47ac10b-58cc-4372-a567-0e02b2c3d002"
	validUUID       = "f47ac10b-58cc-4372-a567-0e02b2c3d009"

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
