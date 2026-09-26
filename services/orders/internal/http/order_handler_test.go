package http

import (
	"bytes"
	"commerce-platform/services/orders/internal/grpc"
	"commerce-platform/services/orders/internal/order"
	"commerce-platform/services/orders/internal/repository"
	"commerce-platform/services/orders/internal/service"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
func setupOrderHandlerTest(t *testing.T) (*httptest.Server, *repository.InMemoryOrderRepository) {
	t.Helper()
	repo := repository.NewInMemoryOrderRepository()
	client := &mockProductsClient{
		productIDs: map[string]bool{
			repository.FirstProductID:  true,
			repository.SecondProductID: true,
		},
	}
	svc := service.NewOrderService(repo, client)
	handler := NewOrderHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	return srv, repo
}

// To be used as BeforeEach
func setupOrderMuxHandlerTest(t *testing.T) (*chi.Mux, *repository.InMemoryOrderRepository) {
	t.Helper()
	repo := repository.NewInMemoryOrderRepository()
	client := &mockProductsClient{
		productIDs: map[string]bool{
			repository.FirstProductID:  true,
			repository.SecondProductID: true,
		},
	}
	svc := service.NewOrderService(repo, client)
	handler := NewOrderHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	return r, repo
}

func TestGetOrders_WhenOrdersExist_Returns200(t *testing.T) {
	srv, _ := setupOrderHandlerTest(t)

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
			"id":         repository.FirstOrderID.String(),
			"product_id": repository.FirstProductID,
			"quantity":   float64(2),
			"status":     "CREATED",
		},
		{
			"id":         repository.SecondOrderID.String(),
			"product_id": repository.SecondProductID,
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

func TestGetOrder_WhenDbErrorHappens_Returns500(t *testing.T) {
	r, _ := setupOrderMuxHandlerTest(t)

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
	srv, _ := setupOrderHandlerTest(t)

	res, err := http.Get(srv.URL + "/orders/" + repository.FirstOrderID.String())

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	var existing order.Order
	err = json.Unmarshal(body, &existing)
	assert.NoError(t, err)
	assert.Equal(t, repository.FirstOrderID, existing.ID)
	assert.Equal(t, repository.FirstProductID, existing.ProductID)
	assert.Equal(t, 2, existing.Quantity)
	assert.Equal(t, order.CREATED, existing.Status)
	assert.False(t, existing.CreatedAt.IsZero())
}

func TestGetOrder_WhenOrderNotExists_Returns404(t *testing.T) {
	srv, _ := setupOrderHandlerTest(t)
	id, _ := uuid.NewV7()

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

func TestCreateOrder_WhenRequestValid_CreatesOrder(t *testing.T) {
	srv, repo := setupOrderHandlerTest(t)

	res, err := http.Post(
		srv.URL+"/orders",
		"application/json",
		bytes.NewBufferString(`{
			"product_id": "`+repository.FirstProductID+`",
			"quantity": 1
		}`),
	)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusCreated, res.StatusCode)

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	// decode response to get the server-assigned ID
	var created order.Order
	err = json.Unmarshal(body, &created)
	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID) // a UUID was assigned
	assert.Equal(t, repository.FirstProductID, created.ProductID)
	assert.Equal(t, 1, created.Quantity)
	assert.Equal(t, order.CREATED, created.Status)
	assert.WithinDuration(t, time.Now(), created.CreatedAt, time.Second)
	assert.Equal(t, "/orders/"+created.ID.String(), res.Header.Get("Location"))

	// verify it was actually persisted
	p, err := repo.FindByID(context.Background(), created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, p.ID)
	assert.Equal(t, repository.FirstProductID, p.ProductID)
	assert.Equal(t, 1, p.Quantity)
	assert.Equal(t, order.CREATED, p.Status)
	assert.WithinDuration(t, time.Now(), p.CreatedAt, time.Second)
}

func TestCreateOrder_WhenProductNotExists_Returns409(t *testing.T) {
	srv, repo := setupOrderHandlerTest(t)

	res, err := http.Post(
		srv.URL+"/orders",
		"application/json",
		bytes.NewBufferString(`{
			"product_id": "999",
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

	orders, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, orders, 2)
}

func TestCreateOrder_WhenBadRequestBody_Returns400(t *testing.T) {
	srv, _ := setupOrderHandlerTest(t)

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
}

func TestCreateOrder_WhenRequestInvalid_Returns400(t *testing.T) {
	srv, _ := setupOrderHandlerTest(t)

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
}

func TestUpdateOrder_WhenRequestValid_UpdatesOrder(t *testing.T) {
	srv, repo := setupOrderHandlerTest(t)

	p, err := repo.FindByID(context.Background(), repository.FirstOrderID)
	assert.NoError(t, err)
	assert.Equal(t, repository.FirstOrderID, p.ID)
	assert.Equal(t, repository.FirstProductID, p.ProductID)
	assert.Equal(t, 2, p.Quantity)
	assert.Equal(t, order.CREATED, p.Status)
	assert.WithinDuration(t, time.Now(), p.CreatedAt, time.Second)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+repository.FirstOrderID.String(),
		bytes.NewBufferString(`{
			"product_id": "`+repository.FirstProductID+`",
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

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	var updated order.Order
	err = json.Unmarshal(body, &updated)
	assert.NoError(t, err)
	assert.Equal(t, repository.FirstOrderID, updated.ID)
	assert.Equal(t, repository.FirstProductID, updated.ProductID)
	assert.Equal(t, 2, updated.Quantity)
	assert.Equal(t, order.PAID, updated.Status)
	assert.True(t, updated.CreatedAt.Equal(p.CreatedAt))

	p, err = repo.FindByID(context.Background(), repository.FirstOrderID)
	assert.NoError(t, err)
	assert.Equal(t, repository.FirstOrderID, p.ID)
	assert.Equal(t, repository.FirstProductID, p.ProductID)
	assert.Equal(t, 2, p.Quantity)
	assert.Equal(t, order.PAID, p.Status)
}

func TestUpdateOrder_WhenRequestValidWithLowercaseStatus_UpdatesOrder(t *testing.T) {
	srv, repo := setupOrderHandlerTest(t)

	p, err := repo.FindByID(context.Background(), repository.FirstOrderID)
	assert.NoError(t, err)
	assert.Equal(t, repository.FirstOrderID, p.ID)
	assert.Equal(t, repository.FirstProductID, p.ProductID)
	assert.Equal(t, 2, p.Quantity)
	assert.Equal(t, order.CREATED, p.Status)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+repository.FirstOrderID.String(),
		bytes.NewBufferString(`{
			"product_id": "`+repository.FirstProductID+`",
			"quantity": 2,
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
	assert.Equal(t, repository.FirstOrderID, updated.ID)
	assert.Equal(t, repository.FirstProductID, updated.ProductID)
	assert.Equal(t, 2, updated.Quantity)
	assert.Equal(t, order.PAID, updated.Status)

	p, err = repo.FindByID(context.Background(), repository.FirstOrderID)
	assert.NoError(t, err)
	assert.Equal(t, repository.FirstOrderID, p.ID)
	assert.Equal(t, repository.FirstProductID, p.ProductID)
	assert.Equal(t, 2, p.Quantity)
	assert.Equal(t, order.PAID, p.Status)
}

func TestUpdateOrder_WhenProductNotExists_Returns409(t *testing.T) {
	srv, repo := setupOrderHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+repository.FirstOrderID.String(),
		bytes.NewBufferString(`{
			"product_id": "999",
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

	// order unchanged
	p, err := repo.FindByID(context.Background(), repository.FirstOrderID)
	assert.NoError(t, err)
	assert.Equal(t, repository.FirstProductID, p.ProductID)
	assert.Equal(t, order.CREATED, p.Status)
}

func TestUpdateOrder_WhenBadRequestBody_Returns400(t *testing.T) {
	srv, _ := setupOrderHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+repository.FirstOrderID.String(),
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
}

func TestUpdateOrder_WhenRequestInvalid_Returns400(t *testing.T) {
	srv, _ := setupOrderHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+repository.FirstOrderID.String(),
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
}

func TestUpdateOrder_WhenBadUUID_Returns400(t *testing.T) {
	srv, _ := setupOrderHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/1234",
		bytes.NewBufferString(`{
			"product_id": "`+repository.FirstProductID+`",
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
}

func TestUpdateOrder_WhenOrderNotExists_Returns404(t *testing.T) {
	srv, repo := setupOrderHandlerTest(t)
	id, _ := uuid.NewV7()

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/orders/"+id.String(),
		bytes.NewBufferString(`{
			"product_id": "`+repository.FirstProductID+`",
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

	_, err = repo.FindByID(context.Background(), id)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestDeleteOrder_WhenOrderExists_DeletesOrder(t *testing.T) {
	srv, repo := setupOrderHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodDelete,
		srv.URL+"/orders/"+repository.SecondOrderID.String(),
		nil,
	)
	assert.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	assert.Equal(t, http.StatusNoContent, res.StatusCode)

	_, err = repo.FindByID(context.Background(), repository.SecondOrderID)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestDeleteOrder_WhenBadUUID_Returns400(t *testing.T) {
	srv, repo := setupOrderHandlerTest(t)

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

	orders, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, orders, 2)
}

func TestDeleteOrder_WhenDbErrorHappens_Returns500(t *testing.T) {
	r, repo := setupOrderMuxHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/orders/"+repository.SecondOrderID.String(),
		nil,
	)
	req = req.WithContext(context.WithValue(req.Context(), "errorEnabler", "unexpected error"))
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)

	_, err := repo.FindByID(context.Background(), repository.SecondOrderID)
	assert.NoError(t, err) // The order should still exist because the delete failed
}
