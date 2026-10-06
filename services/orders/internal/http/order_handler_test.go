package http

import (
	"commerce-platform/services/orders/internal/order"
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
	"github.com/stretchr/testify/assert"
)

func setupOrderHandlerTest(t *testing.T) (*httptest.Server, *fakeOrderService) {
	t.Helper()
	svc := newFakeOrderService(t)
	handler := NewOrderHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	return srv, svc
}

func setupOrderMuxHandlerTest(t *testing.T) (*chi.Mux, *fakeOrderService) {
	t.Helper()
	svc := newFakeOrderService(t)
	handler := NewOrderHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	return r, svc
}

func TestGetOrders_WhenOrdersExist_Returns200(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
	mock.expectGetOrders(
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
}

func TestGetOrders_WhenDbErrorHappens_Returns500(t *testing.T) {
	r, mock := setupOrderMuxHandlerTest(t)
	mock.expectGetOrders(
		func(_ context.Context) ([]order.Order, error) {
			return nil, errors.New("database unavailable")
		})

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
}

func TestGetOrder_WhenOrderExists_Returns200(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
	mock.expectGetOrderByID(
		func(_ context.Context, id uuid.UUID) (order.Order, error) {
			if id == testO.ID {
				return testO, nil
			} else {
				return order.Order{}, errors.New("wrong id passed in fake")
			}
		})

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
}

func TestGetOrder_WhenOrderNotExists_Returns404(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
	id, _ := uuid.NewV7()
	mock.expectGetOrderByID(
		func(_ context.Context, id uuid.UUID) (order.Order, error) {
			return order.Order{}, &service.ErrOrderNotFound{OrderID: id}
		})

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
}

func TestGetOrder_WhenBadUUID_Returns400(t *testing.T) {
	srv, _ := setupOrderHandlerTest(t)

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
}

// func TestCreateOrder_WhenRequestValid_CreatesOrder(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)
// 	mock.ExpectQuery(`
// 			INSERT INTO orders (order_id, product_id, quantity, status)
// 			VALUES ($1, $2, $3, $4)
// 			RETURNING order_id, product_id, quantity, status, created_at;`).
// 		WithArgs(pgxmock.OfType[uuid.UUID](), FirstProductID, 10, order.CREATED).
// 		WillReturnRows(
// 			pgxmock.NewRows([]string{
// 				"order_id", "product_id", "quantity", "status", "created_at",
// 			}).
// 				AddRow(uuid.New(), FirstProductID, 10, order.CREATED, time.Now()),
// 		)

// 	res, err := http.Post(
// 		srv.URL+"/orders",
// 		"application/json",
// 		bytes.NewBufferString(`{
// 			"product_id": " `+FirstProductID+` ",
// 			"quantity": 10
// 		}`),
// 	)

// 	assert.NoError(t, err)
// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	assert.Equal(t, http.StatusCreated, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
// 	var created order.Order
// 	err = json.Unmarshal(body, &created)
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, created.ID) // a UUID was assigned
// 	assert.Equal(t, FirstProductID, created.ProductID)
// 	assert.Equal(t, 10, created.Quantity)
// 	assert.Equal(t, order.CREATED, created.Status)
// 	assert.WithinDuration(t, time.Now(), created.CreatedAt, time.Second)
// 	assert.Equal(t, "/orders/"+created.ID.String(), res.Header.Get("Location"))
// }

// func TestCreateOrder_WhenRequestBodyTooBig_Returns400(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)

// 	req, err := http.NewRequest(
// 		http.MethodPost,
// 		srv.URL+"/orders",
// 		bytes.NewBufferString(`{
// 			"product_id": "`+FirstProductID+`",
// 			"quantity": 11,
// 			"status": "`+strings.Repeat("a", 1<<10)+`"
// 		}{"trailing": "json"}`),
// 	)
// 	assert.NoError(t, err)
// 	req.Header.Set("Content-Type", "application/json")
// 	res, err := http.DefaultClient.Do(req)

// 	assert.NoError(t, err)
// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
// 	assert.JSONEq(
// 		t,
// 		`{
// 			"code": "REQUEST_BODY_TOO_LONG",
// 			"message": "Request body exceeds 1024 bytes"
// 		}`,
// 		string(body),
// 	)
// }

// func TestCreateOrder_WhenRequestBodyWithTrailingJson_Returns400(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)

// 	req, err := http.NewRequest(
// 		http.MethodPost,
// 		srv.URL+"/orders",
// 		bytes.NewBufferString(`{
// 			"product_id": "`+FirstProductID+`",
// 			"quantity": 11,
// 			"status": "paid"
// 		}{"trailing": "json"}`),
// 	)
// 	assert.NoError(t, err)
// 	req.Header.Set("Content-Type", "application/json")
// 	res, err := http.DefaultClient.Do(req)

// 	assert.NoError(t, err)
// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
// 	assert.JSONEq(
// 		t,
// 		`{
// 			"code": "INVALID_ORDER",
// 			"message": "invalid order"
// 		}`,
// 		string(body),
// 	)
// }

// func TestCreateOrder_WhenProductNotExists_Returns409(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)

// 	res, err := http.Post(
// 		srv.URL+"/orders",
// 		"application/json",
// 		bytes.NewBufferString(`{
// 			"product_id": "`+validUUID+`",
// 			"quantity": 1
// 		}`),
// 	)

// 	assert.NoError(t, err)
// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	assert.Equal(t, http.StatusConflict, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
// 	assert.JSONEq(
// 		t,
// 		`{
// 			"code": "PRODUCT_NOT_FOUND",
// 			"message": "product not found for the given id"
// 		}`,
// 		string(body),
// 	)
// }

// func TestCreateOrder_WhenBadRequestBody_Returns400(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)

// 	res, err := http.Post(
// 		srv.URL+"/orders",
// 		"application/json",
// 		bytes.NewBufferString(`{
// 			"error-to-cause": "extra comma, so invalid json",
// 		}`),
// 	)

// 	assert.NoError(t, err)
// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
// 	assert.JSONEq(
// 		t,
// 		`{
// 			"code": "INVALID_ORDER",
// 			"message": "invalid order"
// 		}`,
// 		string(body),
// 	)
// }

// func TestCreateOrder_WhenRequestInvalid_Returns400(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)

// 	res, err := http.Post(
// 		srv.URL+"/orders",
// 		"application/json",
// 		bytes.NewBufferString(`{
// 			"product_id": "",
// 			"quantity": 0
// 		}`),
// 	)

// 	assert.NoError(t, err)
// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
// 	assert.JSONEq(
// 		t,
// 		`{
// 			"code": "VALIDATION_ERROR",
// 			"message": "product-id cannot be blank.; quantity must be > 0."
// 		}`,
// 		string(body),
// 	)
// }

// func TestUpdateOrder_WhenRequestValid_UpdatesOrder(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)
// 	mock.ExpectQuery(`
// 			SELECT order_id, product_id, quantity, status, created_at
// 			FROM orders
// 			WHERE order_id = $1`).
// 		WithArgs(testO.ID).
// 		WillReturnRows(
// 			pgxmock.NewRows([]string{
// 				"order_id", "product_id", "quantity", "status", "created_at",
// 			}).
// 				AddRow(testO.ID, testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt),
// 		)
// 	mock.ExpectQuery(`
// 			UPDATE orders
// 			SET product_id = $1, quantity = $2, status = $3
// 			WHERE order_id = $4
// 			RETURNING order_id, product_id, quantity, status, created_at;`).
// 		WithArgs(FirstProductID, 11, order.PAID, testO.ID).
// 		WillReturnRows(
// 			pgxmock.NewRows([]string{
// 				"order_id", "product_id", "quantity", "status", "created_at",
// 			}).
// 				AddRow(testO.ID, FirstProductID, 11, order.PAID, testO.CreatedAt),
// 		)

// 	req, err := http.NewRequest(
// 		http.MethodPut,
// 		srv.URL+"/orders/"+FirstOrderID.String(),
// 		bytes.NewBufferString(`{
// 			"product_id": " `+FirstProductID+`",
// 			"quantity": 11,
// 			"status": "paid"
// 		}`),
// 	)
// 	assert.NoError(t, err)
// 	req.Header.Set("Content-Type", "application/json")
// 	res, err := http.DefaultClient.Do(req)

// 	assert.NoError(t, err)
// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	assert.Equal(t, http.StatusOK, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
// 	var updated order.Order
// 	err = json.Unmarshal(body, &updated)
// 	assert.NoError(t, err)
// 	assert.Equal(t, testO.ID, updated.ID)
// 	assert.Equal(t, FirstProductID, updated.ProductID)
// 	assert.Equal(t, 11, updated.Quantity)
// 	assert.Equal(t, order.PAID, updated.Status)
// 	assert.True(t, updated.CreatedAt.Equal(testO.CreatedAt))
// }

// func TestUpdateOrder_WhenRequestValidWithLowercaseStatus_UpdatesOrder(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)
// 	mock.ExpectQuery(`
// 			SELECT order_id, product_id, quantity, status, created_at
// 			FROM orders
// 			WHERE order_id = $1`).
// 		WithArgs(testO.ID).
// 		WillReturnRows(
// 			pgxmock.NewRows([]string{
// 				"order_id", "product_id", "quantity", "status", "created_at",
// 			}).
// 				AddRow(testO.ID, testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt),
// 		)
// 	mock.ExpectQuery(`
// 			UPDATE orders
// 			SET product_id = $1, quantity = $2, status = $3
// 			WHERE order_id = $4
// 			RETURNING order_id, product_id, quantity, status, created_at;`).
// 		WithArgs(FirstProductID, 11, order.PAID, testO.ID).
// 		WillReturnRows(
// 			pgxmock.NewRows([]string{
// 				"order_id", "product_id", "quantity", "status", "created_at",
// 			}).
// 				AddRow(testO.ID, FirstProductID, 11, order.PAID, testO.CreatedAt),
// 		)

// 	req, err := http.NewRequest(
// 		http.MethodPut,
// 		srv.URL+"/orders/"+FirstOrderID.String(),
// 		bytes.NewBufferString(`{
// 			"product_id": "`+FirstProductID+`",
// 			"quantity": 11,
// 			"status": "paid"
// 		}`),
// 	)
// 	assert.NoError(t, err)
// 	req.Header.Set("Content-Type", "application/json")
// 	res, err := http.DefaultClient.Do(req)

// 	assert.NoError(t, err)
// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	assert.Equal(t, http.StatusOK, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
// 	var updated order.Order
// 	err = json.Unmarshal(body, &updated)
// 	assert.NoError(t, err)
// 	assert.Equal(t, testO.ID, updated.ID)
// 	assert.Equal(t, FirstProductID, updated.ProductID)
// 	assert.Equal(t, 11, updated.Quantity)
// 	assert.Equal(t, order.PAID, updated.Status)
// 	assert.True(t, updated.CreatedAt.Equal(testO.CreatedAt))
// }

// func TestUpdateOrder_WhenRequestBodyTooBig_Returns400(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)

// 	req, err := http.NewRequest(
// 		http.MethodPut,
// 		srv.URL+"/orders/"+FirstOrderID.String(),
// 		bytes.NewBufferString(`{
// 			"product_id": "`+FirstProductID+`",
// 			"quantity": 11,
// 			"status": "`+strings.Repeat("a", 1<<10)+`"
// 		}{"trailing": "json"}`),
// 	)
// 	assert.NoError(t, err)
// 	req.Header.Set("Content-Type", "application/json")
// 	res, err := http.DefaultClient.Do(req)

// 	assert.NoError(t, err)
// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
// 	assert.JSONEq(
// 		t,
// 		`{
// 			"code": "REQUEST_BODY_TOO_LONG",
// 			"message": "Request body exceeds 1024 bytes"
// 		}`,
// 		string(body),
// 	)
// }

// func TestUpdateOrder_WhenRequestBodyWithTrailingJson_Returns400(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)

// 	req, err := http.NewRequest(
// 		http.MethodPut,
// 		srv.URL+"/orders/"+FirstOrderID.String(),
// 		bytes.NewBufferString(`{
// 			"product_id": "`+FirstProductID+`",
// 			"quantity": 11,
// 			"status": "paid"
// 		}{"trailing": "json"}`),
// 	)
// 	assert.NoError(t, err)
// 	req.Header.Set("Content-Type", "application/json")
// 	res, err := http.DefaultClient.Do(req)

// 	assert.NoError(t, err)
// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
// 	assert.JSONEq(
// 		t,
// 		`{
// 			"code": "INVALID_ORDER",
// 			"message": "invalid order"
// 		}`,
// 		string(body),
// 	)
// }

// func TestUpdateOrder_WhenProductNotExists_Returns409(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)
// 	mock.ExpectQuery(`
// 			SELECT order_id, product_id, quantity, status, created_at
// 			FROM orders
// 			WHERE order_id = $1`).
// 		WithArgs(testO.ID).
// 		WillReturnRows(
// 			pgxmock.NewRows([]string{
// 				"order_id", "product_id", "quantity", "status", "created_at",
// 			}).
// 				AddRow(testO.ID, testO.ProductID, testO.Quantity, testO.Status, testO.CreatedAt),
// 		)

// 	req, err := http.NewRequest(
// 		http.MethodPut,
// 		srv.URL+"/orders/"+testO.ID.String(),
// 		bytes.NewBufferString(`{
// 			"product_id": "`+validUUID+`",
// 			"quantity": 2,
// 			"status": "PAID"
// 		}`),
// 	)
// 	assert.NoError(t, err)
// 	req.Header.Set("Content-Type", "application/json")
// 	res, err := http.DefaultClient.Do(req)

// 	assert.NoError(t, err)
// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	assert.Equal(t, http.StatusConflict, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
// 	assert.JSONEq(
// 		t,
// 		`{
// 			"code": "PRODUCT_NOT_FOUND",
// 			"message": "product not found for the given id"
// 		}`,
// 		string(body),
// 	)
// }

// func TestUpdateOrder_WhenBadRequestBody_Returns400(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)

// 	req, err := http.NewRequest(
// 		http.MethodPut,
// 		srv.URL+"/orders/"+FirstOrderID.String(),
// 		bytes.NewBufferString(`{
// 			"error-to-cause": "extra comma, so invalid json",
// 		}`),
// 	)
// 	assert.NoError(t, err)
// 	req.Header.Set("Content-Type", "application/json")
// 	res, err := http.DefaultClient.Do(req)

// 	assert.NoError(t, err)
// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
// 	assert.JSONEq(
// 		t,
// 		`{
// 			"code": "INVALID_ORDER",
// 			"message": "invalid order"
// 		}`,
// 		string(body),
// 	)
// }

// func TestUpdateOrder_WhenRequestInvalid_Returns400(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)

// 	req, err := http.NewRequest(
// 		http.MethodPut,
// 		srv.URL+"/orders/"+FirstOrderID.String(),
// 		bytes.NewBufferString(`{
// 			"product_id": "",
// 			"quantity": 0,
// 			"status": "PIAD"
// 		}`),
// 	)
// 	assert.NoError(t, err)
// 	req.Header.Set("Content-Type", "application/json")
// 	res, err := http.DefaultClient.Do(req)

// 	assert.NoError(t, err)
// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
// 	assert.JSONEq(
// 		t,
// 		`{
// 			"code": "VALIDATION_ERROR",
// 			"message": "product-id cannot be blank.; quantity must be > 0.; status is not valid."
// 		}`,
// 		string(body),
// 	)
// }

// func TestUpdateOrder_WhenBadUUID_Returns400(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)

// 	req, err := http.NewRequest(
// 		http.MethodPut,
// 		srv.URL+"/orders/1234",
// 		bytes.NewBufferString(`{
// 			"product_id": "`+FirstProductID+`",
// 			"quantity": 1,
// 			"status": "PAID"
// 		}`),
// 	)
// 	assert.NoError(t, err)
// 	req.Header.Set("Content-Type", "application/json")
// 	res, err := http.DefaultClient.Do(req)

// 	assert.NoError(t, err)
// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
// 	assert.JSONEq(
// 		t,
// 		`{
// 			"code": "INVALID_UUID",
// 			"message": "invalid UUID"
// 		}`,
// 		string(body),
// 	)
// }

// func TestUpdateOrder_WhenOrderNotExists_Returns404(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)
// 	id, _ := uuid.NewV7()
// 	mock.ExpectQuery(`
// 			SELECT order_id, product_id, quantity, status, created_at
// 			FROM orders
// 			WHERE order_id = $1`).
// 		WithArgs(id).
// 		WillReturnRows(
// 			pgxmock.NewRows([]string{
// 				"order_id", "product_id", "quantity", "status", "created_at",
// 			}))

// 	req, err := http.NewRequest(
// 		http.MethodPut,
// 		srv.URL+"/orders/"+id.String(),
// 		bytes.NewBufferString(`{
// 			"product_id": "`+FirstProductID+`",
// 			"quantity": 1,
// 			"status": "PAiD"
// 		}`),
// 	)
// 	assert.NoError(t, err)
// 	req.Header.Set("Content-Type", "application/json")
// 	res, err := http.DefaultClient.Do(req)

// 	assert.NoError(t, err)
// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	assert.Equal(t, http.StatusNotFound, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
// 	assert.JSONEq(
// 		t,
// 		`{
// 			"code": "ORDER_NOT_FOUND",
// 			"message": "Order with id [`+id.String()+`] was not found"
// 		}`,
// 		string(body),
// 	)
// }

// func TestDeleteOrder_WhenOrderExists_DeletesOrder(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)
// 	mock.ExpectExec(`
// 			DELETE FROM orders
// 			WHERE order_id = $1`).
// 		WithArgs(SecondOrderID).
// 		WillReturnResult(pgxmock.NewResult("DELETE", 1))

// 	req, err := http.NewRequest(
// 		http.MethodDelete,
// 		srv.URL+"/orders/"+SecondOrderID.String(),
// 		nil,
// 	)
// 	assert.NoError(t, err)
// 	res, err := http.DefaultClient.Do(req)

// 	assert.Equal(t, http.StatusNoContent, res.StatusCode)
// }

// func TestDeleteOrder_WhenBadUUID_Returns400(t *testing.T) {
// 	srv, mock := setupOrderHandlerTest(t)

// 	req, err := http.NewRequest(
// 		http.MethodDelete,
// 		srv.URL+"/orders/1234",
// 		nil,
// 	)
// 	assert.NoError(t, err)
// 	res, err := http.DefaultClient.Do(req)

// 	assert.NoError(t, err)
// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
// 	assert.JSONEq(
// 		t,
// 		`{
// 			"code": "INVALID_UUID",
// 			"message": "invalid UUID"
// 		}`,
// 		string(body),
// 	)
// }

// func TestDeleteOrder_WhenDbErrorHappens_Returns500(t *testing.T) {
// 	r, mock := setupOrderMuxHandlerTest(t)
// 	mock.ExpectExec(`
// 			DELETE FROM orders
// 			WHERE order_id = $1`).
// 		WithArgs(SecondOrderID).
// 		WillReturnError(errors.New("database unavailable"))

// 	req := httptest.NewRequest(
// 		http.MethodDelete,
// 		"/orders/"+SecondOrderID.String(),
// 		nil,
// 	)
// 	req = req.WithContext(context.WithValue(req.Context(), "errorEnabler", "unexpected error"))
// 	res := httptest.NewRecorder()

// 	r.ServeHTTP(res, req)

// 	assert.Equal(t, http.StatusInternalServerError, res.Code)
// 	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
// 	body, _ := io.ReadAll(res.Body)
// 	assert.JSONEq(
// 		t,
// 		`{
// 			"code": "INTERNAL_SERVER_ERROR",
// 			"message": "internal server error"
// 		}`,
// 		string(body),
// 	)
// }

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
