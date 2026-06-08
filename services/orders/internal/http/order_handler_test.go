package http

import (
	"commerce-platform/services/orders/internal/repository"
	"commerce-platform/services/orders/internal/service"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestGetOrders_WhenOrdersExist_Returns200(t *testing.T) {
	repo := repository.NewInMemoryOrderRepository()
	svc := service.NewOrderService(repo)
	handler := NewOrderHandler(svc)

	r := chi.NewRouter()

	handler.RegisterRoutes(r)

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
	repo := repository.NewInMemoryOrderRepository()
	svc := service.NewOrderService(repo)
	handler := NewOrderHandler(svc)

	r := chi.NewRouter()

	handler.RegisterRoutes(r)

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
	repo := repository.NewInMemoryOrderRepository()
	svc := service.NewOrderService(repo)
	handler := NewOrderHandler(svc)

	r := chi.NewRouter()

	handler.RegisterRoutes(r)

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
