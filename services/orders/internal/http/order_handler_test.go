package http

import (
	"bytes"
	"commerce-platform/services/orders/internal/grpc"
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/services/orders/internal/repository"
	"commerce-platform/services/orders/internal/service"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
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
func setupOrderHandlerTest(t *testing.T) (*chi.Mux, *repository.InMemoryOrderRepository) {
	t.Helper()
	repo := repository.NewInMemoryOrderRepository()
	client := &mockProductsClient{productIDs: map[string]bool{"1": true, "2": true, "3": true}}
	svc := service.NewOrderService(repo, client)
	handler := NewOrderHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	return r, repo
}

func TestGetOrders_WhenOrdersExist_Returns200(t *testing.T) {
	r, _ := setupOrderHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders",
		nil,
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))

	var resOrders []map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &resOrders)
	assert.NoError(t, err)

	expectedOrders := []map[string]any{
		{
			"id":         "1",
			"product_id": "1",
			"quantity":   float64(2),
			"status":     "CREATED",
		},
		{
			"id":         "2",
			"product_id": "2",
			"quantity":   float64(1),
			"status":     "PAID",
		},
	}

	assert.ElementsMatch(t, expectedOrders, resOrders)
}

func TestGetOrder_WhenOrderExists_Returns200(t *testing.T) {
	r, _ := setupOrderHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders/1",
		nil,
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"id":         "1",
			"product_id": "1",
			"quantity":   2,
			"status":     "CREATED"
		}`,
		res.Body.String(),
	)
}

func TestGetOrder_WhenOrderNotExists_Returns404(t *testing.T) {
	r, _ := setupOrderHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders/6",
		nil,
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "ORDER_NOT_FOUND",
			"message": "order not found"
		}`,
		res.Body.String(),
	)
}

func TestCreateOrder_WhenRequestValid_CreatesOrder(t *testing.T) {
	r, repo := setupOrderHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders",
		bytes.NewBufferString(`{
			"id": "3",
			"product_id": "1",
			"quantity": 1
		}`),
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusCreated, res.Code)

	p, exists := repo.FindByID("3")
	assert.True(t, exists)
	assert.Equal(t, "3", p.ID)
	assert.Equal(t, "1", p.ProductID)
	assert.Equal(t, 1, p.Quantity)
}

func TestCreateOrder_WhenProductNotExists_Returns409(t *testing.T) {
	r, repo := setupOrderHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders",
		bytes.NewBufferString(`{
			"id": "3",
			"product_id": "999",
			"quantity": 1
		}`),
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusConflict, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "PRODUCT_NOT_FOUND",
			"message": "product not found for the given id"
		}`,
		res.Body.String(),
	)

	_, exists := repo.FindByID("3")
	assert.False(t, exists)
}

func TestCreateOrder_WhenBadRequestBody_Returns400(t *testing.T) {
	r, _ := setupOrderHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders",
		bytes.NewBufferString(`{
			"error-to-cause": "extra comma, so invalid json",
		}`),
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INVALID_ORDER",
			"message": "invalid order"
		}`,
		res.Body.String(),
	)
}

func TestCreateOrder_WhenRequestInvalid_Returns400(t *testing.T) {
	r, repo := setupOrderHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders",
		bytes.NewBufferString(`{
			"id": "",
			"product_id": "",
			"quantity": 0
		}`),
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "VALIDATION_ERROR",
			"message": "id cannot be blank.; product-id cannot be blank.; quantity must be > 0."
		}`,
		res.Body.String(),
	)

	_, exists := repo.FindByID("")
	assert.False(t, exists)
}

func TestUpdateOrder_WhenRequestValid_UpdatesOrder(t *testing.T) {
	r, repo := setupOrderHandlerTest(t)

	p, exists := repo.FindByID("1")
	assert.True(t, exists)
	assert.Equal(t, "1", p.ID)
	assert.Equal(t, "1", p.ProductID)
	assert.Equal(t, 2, p.Quantity)
	assert.Equal(t, order.CREATED, p.Status)

	req := httptest.NewRequest(
		http.MethodPut,
		"/orders/1",
		bytes.NewBufferString(`{
			"product_id": "1",
			"quantity": 2,
			"status": "PAID"
		}`),
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)

	p, exists = repo.FindByID("1")
	assert.True(t, exists)
	assert.Equal(t, "1", p.ID)
	assert.Equal(t, "1", p.ProductID)
	assert.Equal(t, 2, p.Quantity)
	assert.Equal(t, order.PAID, p.Status)
}

func TestUpdateOrder_WhenRequestValidWithLowercaseStatus_UpdatesOrder(t *testing.T) {
	r, repo := setupOrderHandlerTest(t)

	p, exists := repo.FindByID("1")
	assert.True(t, exists)
	assert.Equal(t, "1", p.ID)
	assert.Equal(t, "1", p.ProductID)
	assert.Equal(t, 2, p.Quantity)
	assert.Equal(t, order.CREATED, p.Status)

	req := httptest.NewRequest(
		http.MethodPut,
		"/orders/1",
		bytes.NewBufferString(`{
			"product_id": "1",
			"quantity": 2,
			"status": "paid"
		}`),
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)

	p, exists = repo.FindByID("1")
	assert.True(t, exists)
	assert.Equal(t, "1", p.ID)
	assert.Equal(t, "1", p.ProductID)
	assert.Equal(t, 2, p.Quantity)
	assert.Equal(t, order.PAID, p.Status)
}

func TestUpdateOrder_WhenProductNotExists_Returns409(t *testing.T) {
	r, repo := setupOrderHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPut,
		"/orders/1",
		bytes.NewBufferString(`{
			"product_id": "999",
			"quantity": 2,
			"status": "PAID"
		}`),
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusConflict, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "PRODUCT_NOT_FOUND",
			"message": "product not found for the given id"
		}`,
		res.Body.String(),
	)

	// order unchanged
	p, exists := repo.FindByID("1")
	assert.True(t, exists)
	assert.Equal(t, "1", p.ProductID)
	assert.Equal(t, order.CREATED, p.Status)
}

func TestUpdateOrder_WhenBadRequestBody_Returns400(t *testing.T) {
	r, _ := setupOrderHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPut,
		"/orders/1",
		bytes.NewBufferString(`{
			"error-to-cause": "extra comma, so invalid json",
		}`),
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INVALID_ORDER",
			"message": "invalid order"
		}`,
		res.Body.String(),
	)
}

func TestUpdateOrder_WhenRequestInvalid_Returns400(t *testing.T) {
	r, repo := setupOrderHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPut,
		"/orders/1",
		bytes.NewBufferString(`{
			"product_id": "",
			"quantity": 0,
			"status": "PIAD"
		}`),
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "VALIDATION_ERROR",
			"message": "product-id cannot be blank.; quantity must be > 0.; status is not valid."
		}`,
		res.Body.String(),
	)

	_, exists := repo.FindByID("")
	assert.False(t, exists)
}

func TestUpdateOrder_WhenOrderNotExists_Returns404(t *testing.T) {
	r, repo := setupOrderHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPut,
		"/orders/11",
		bytes.NewBufferString(`{
			"product_id": "1",
			"quantity": 1,
			"status": "PAiD"
		}`),
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "ORDER_NOT_FOUND",
			"message": "order not found"
		}`,
		res.Body.String(),
	)

	_, exists := repo.FindByID("11")
	assert.False(t, exists)
}

func TestDeleteOrder_WhenOrderExists_DeletesOrder(t *testing.T) {
	r, repo := setupOrderHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/orders/2",
		nil,
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNoContent, res.Code)

	_, exists := repo.FindByID("2")
	assert.False(t, exists)
}
