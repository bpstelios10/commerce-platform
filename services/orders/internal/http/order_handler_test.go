package http

import (
	"bytes"
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/services/orders/internal/service"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
					Version:   1,
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
}

func TestGetOrders_WhenDbErrorHappens_Returns500(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
	mock.expectGetOrders(
		func(_ context.Context) ([]order.Order, error) {
			return nil, errors.New("database unavailable")
		})

	res, err := http.Get(srv.URL + "/orders")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INTERNAL_SERVER_ERROR",
			"message": "internal server error"
		}`,
		string(body),
	)
}

func TestGetOrder_WhenOrderExists_Returns200(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
	mock.expectGetOrderByID(
		func(_ context.Context, id uuid.UUID) (order.Order, error) {
			return testO, nil
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
	assert.Equal(t, testO.ID, mock.searchedIDs[0])
}

func TestGetOrder_WhenOrderNotExists_Returns404(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
	id, _ := uuid.NewV7()
	mock.expectGetOrderByID(
		func(_ context.Context, id uuid.UUID) (order.Order, error) {
			return order.Order{}, fmt.Errorf("blah: %w", &service.ErrOrderNotFound{OrderID: id})
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
	assert.Equal(t, id, mock.searchedIDs[0])
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
	assert.Empty(t, mock.searchedIDs)
}

func TestCreateOrder_WhenRequestValid_CreatesOrder(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
	mock.expectCreateOrder(
		func(_ context.Context, productID string, quantity int) (order.Order, error) {
			o := order.Order{
				ID:        uuid.New(),
				ProductID: productID,
				Quantity:  quantity,
				Status:    order.CREATED,
				CreatedAt: time.Now(),
			}
			return o, nil
		})

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
	assert.Equal(t, FirstProductID, mock.created[0].ProductID)
	assert.Equal(t, 10, mock.created[0].Quantity)
}

func TestCreateOrder_WhenRequestBodyTooBig_Returns400(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPost,
		srv.URL+"/orders",
		bytes.NewBufferString(`{
			"product_id": "`+FirstProductID+`",
			"quantity": 11,
			"status": "`+strings.Repeat("a", 1<<10)+`"
		}{"trailing": "json"}`),
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
			"code": "REQUEST_BODY_TOO_LONG",
			"message": "Request body exceeds 1024 bytes"
		}`,
		string(body),
	)
	assert.Empty(t, mock.created)
}

func TestCreateOrder_WhenRequestBodyWithTrailingJson_Returns400(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPost,
		srv.URL+"/orders",
		bytes.NewBufferString(`{
			"product_id": "`+FirstProductID+`",
			"quantity": 11,
			"status": "paid"
		}{"trailing": "json"}`),
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
	assert.Empty(t, mock.created)
}

func TestCreateOrder_WhenProductNotExists_Returns409(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
	mock.expectCreateOrder(
		func(_ context.Context, productID string, quantity int) (order.Order, error) {
			return order.Order{}, fmt.Errorf("blah: %w", service.ErrProductNotFound)
		})

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
	assert.Equal(t, validUUID, mock.created[0].ProductID)
	assert.Equal(t, 1, mock.created[0].Quantity)
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
	assert.Empty(t, mock.created)
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
	assert.Empty(t, mock.created)
}

func TestUpdateOrder_WhenRequestValid_UpdatesOrder(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
	mock.expectUpdateOrder(
		func(ctx context.Context, id uuid.UUID, productID string, quantity int, status order.OrderStatus, version int) (order.Order, error) {
			return order.Order{
				ID:        id,
				ProductID: productID,
				Quantity:  quantity,
				Status:    status,
				CreatedAt: testO.CreatedAt,
				Version:   testO.Version + 1,
			}, nil
		})

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
	assert.Equal(t, 2, updated.Version)
	assert.Equal(t, FirstOrderID, mock.updated[0].ID)
	assert.Equal(t, FirstProductID, mock.updated[0].ProductID)
	assert.Equal(t, 11, mock.updated[0].Quantity)
	assert.Equal(t, order.PAID, mock.updated[0].Status)
	assert.Equal(t, 1, mock.updated[0].Version)
}

func TestUpdateOrder_WhenRequestValidWithLowercaseStatus_UpdatesOrder(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
	mock.expectUpdateOrder(
		func(ctx context.Context, id uuid.UUID, productID string, quantity int, status order.OrderStatus, version int) (order.Order, error) {
			return order.Order{
				ID:        id,
				ProductID: productID,
				Quantity:  quantity,
				Status:    status,
				CreatedAt: testO.CreatedAt,
				Version:   1,
			}, nil
		})

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+FirstOrderID.String(),
		bytes.NewBufferString(`{
			"product_id": "`+FirstProductID+`",
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
	assert.Equal(t, FirstOrderID, mock.updated[0].ID)
	assert.Equal(t, FirstProductID, mock.updated[0].ProductID)
	assert.Equal(t, 11, mock.updated[0].Quantity)
	assert.Equal(t, order.PAID, mock.updated[0].Status)
}

func TestUpdateOrder_WhenRequestBodyTooBig_Returns400(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+FirstOrderID.String(),
		bytes.NewBufferString(`{
			"product_id": "`+FirstProductID+`",
			"quantity": 11,
			"status": "`+strings.Repeat("a", 1<<10)+`",
			"version": 1
		}{"trailing": "json"}`),
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
			"code": "REQUEST_BODY_TOO_LONG",
			"message": "Request body exceeds 1024 bytes"
		}`,
		string(body),
	)
	assert.Empty(t, mock.updated)
}

func TestUpdateOrder_WhenRequestBodyWithTrailingJson_Returns400(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+FirstOrderID.String(),
		bytes.NewBufferString(`{
			"product_id": "`+FirstProductID+`",
			"quantity": 11,
			"status": "paid",
			"version": 1
		}{"trailing": "json"}`),
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
	assert.Empty(t, mock.updated)
}

func TestUpdateOrder_WhenProductNotExists_Returns409(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
	mock.expectUpdateOrder(
		func(ctx context.Context, id uuid.UUID, productID string, quantity int, status order.OrderStatus, version int) (order.Order, error) {
			return order.Order{}, fmt.Errorf("blah: %w", service.ErrProductNotFound)
		})

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
	assert.Equal(t, testO.ID, mock.updated[0].ID)
	assert.Equal(t, validUUID, mock.updated[0].ProductID)
	assert.Equal(t, 2, mock.updated[0].Quantity)
	assert.Equal(t, order.PAID, mock.updated[0].Status)
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
	assert.Empty(t, mock.updated)
}

func TestUpdateOrder_WhenRequestInvalid_Returns400(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+FirstOrderID.String(),
		bytes.NewBufferString(`{
			"product_id": "",
			"quantity": 0,
			"status": "PIAD",
			"version": -1
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
			"message": "product-id cannot be blank.; quantity must be > 0.; status is not valid.; version must be > 0."
		}`,
		string(body),
	)
	assert.Empty(t, mock.updated)
}

func TestUpdateOrder_WhenBadUUID_Returns400(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/1234",
		bytes.NewBufferString(`{
			"product_id": "`+FirstProductID+`",
			"quantity": 1,
			"status": "PAID",
			"version": -1
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
	assert.Empty(t, mock.updated)
}

func TestUpdateOrder_WhenUpdateConflict_Returns409(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
	id, _ := uuid.NewV7()
	mock.expectUpdateOrder(
		func(ctx context.Context, id uuid.UUID, productID string, quantity int, status order.OrderStatus, version int) (order.Order, error) {
			return order.Order{}, fmt.Errorf("blah: %w", service.ErrOrderUpdateConflict)
		})

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+id.String(),
		bytes.NewBufferString(`{
			"product_id": "`+FirstProductID+`",
			"quantity": 1,
			"status": "PAiD",
			"version": 2
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
	assert.Equal(t, id, mock.updated[0].ID)
	assert.Equal(t, FirstProductID, mock.updated[0].ProductID)
	assert.Equal(t, 1, mock.updated[0].Quantity)
	assert.Equal(t, order.PAID, mock.updated[0].Status)
}

func TestDeleteOrder_WhenOrderExists_DeletesOrder(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
	mock.expectDeleteOrder(
		func(_ context.Context, id uuid.UUID) error {
			return nil
		})

	req, err := http.NewRequest(
		http.MethodDelete,
		srv.URL+"/orders/"+SecondOrderID.String(),
		nil,
	)
	assert.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, res.StatusCode)
	assert.Equal(t, SecondOrderID, mock.deleted[0])
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
	assert.Empty(t, mock.deleted)
}

func TestDeleteOrder_WhenDbErrorHappens_Returns500(t *testing.T) {
	srv, mock := setupOrderHandlerTest(t)
	mock.expectDeleteOrder(
		func(_ context.Context, id uuid.UUID) error {
			return errors.New("database unavailable")
		})

	req, err := http.NewRequest(
		http.MethodDelete,
		srv.URL+"/orders/"+SecondOrderID.String(),
		nil,
	)
	assert.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INTERNAL_SERVER_ERROR",
			"message": "internal server error"
		}`,
		string(body),
	)
	assert.Equal(t, SecondOrderID, mock.deleted[0])
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
		Version:   1,
	}
)
