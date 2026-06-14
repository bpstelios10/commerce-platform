package httpx

import (
	"bytes"
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupAdminHandlerTest(t *testing.T) (*chi.Mux, *repository.InMemoryProductRepository) {
	t.Helper()
	repo := repository.NewInMemoryProductRepository()
	productService := service.NewProductService(repo)
	adminSvc := service.NewAdminService(productService, repo)
	handler := NewAdminHandler(adminSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	return r, repo
}

func TestGetAdmin_Returns200(t *testing.T) {
	r, _ := setupAdminHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin",
		nil,
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "admin", res.Body.String())
}

func TestCreateProduct_WhenRequestValid_CreatesProduct(t *testing.T) {
	r, repo := setupAdminHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/admin/products",
		bytes.NewBufferString(`{
			"name": "iPad",
			"price": 999
		}`),
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))

	// decode response to get the server-assigned ID
	var created product.Product
	err := json.Unmarshal(res.Body.Bytes(), &created)
	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "iPad", created.Name)
	assert.Equal(t, float64(999), created.Price)
	assert.Equal(t, "/products/"+created.ID.String(), res.Header().Get("Location"))

	// verify it was actually persisted
	p, exists := repo.FindByID(created.ID)
	assert.True(t, exists)
	assert.Equal(t, created.ID, p.ID)
	assert.Equal(t, created.Name, p.Name)
	assert.Equal(t, created.Price, p.Price)
}

func TestCreateProduct_WhenBadRequestBody_Returns400(t *testing.T) {
	r, _ := setupAdminHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/admin/products",
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
			"code": "INVALID_PRODUCT",
			"message": "invalid product"
		}`,
		res.Body.String(),
	)
}

func TestCreateProduct_WhenRequestInvalid_Returns400(t *testing.T) {
	r, _ := setupAdminHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/admin/products",
		bytes.NewBufferString(`{
			"id": "",
			"name": "",
			"price": 0
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
			"message": "name cannot be blank.; price must be > 0."
		}`,
		res.Body.String(),
	)
}

func TestUpdateProduct_WhenRequestValid_UpdatesProduct(t *testing.T) {
	r, repo := setupAdminHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPut,
		"/admin/products/"+repository.SecondUUID.String(),
		bytes.NewBufferString(`{
			"name": "iPhone 15",
			"price": 1500
		}`),
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)

	p, exists := repo.FindByID(repository.SecondUUID)
	assert.True(t, exists)
	assert.Equal(t, repository.SecondUUID, p.ID)
	assert.Equal(t, "iPhone 15", p.Name)
	assert.Equal(t, 1500.0, p.Price)
}

func TestUpdateProduct_WhenBadUUID_Returns400(t *testing.T) {
	r, _ := setupAdminHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPut,
		"/admin/products/1234",
		bytes.NewBufferString(`{
			"name": "iPhone 15",
			"price": 1500
		}`),
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INVALID_UUID",
			"message": "invalid UUID"
		}`,
		res.Body.String(),
	)
}

func TestUpdateProduct_WhenBadRequestBody_Returns400(t *testing.T) {
	r, _ := setupAdminHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPut,
		"/admin/products/"+repository.SecondUUID.String(),
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
			"code": "INVALID_PRODUCT",
			"message": "invalid product"
		}`,
		res.Body.String(),
	)
}

func TestUpdateProduct_WhenRequestInvalid_Returns400(t *testing.T) {
	r, repo := setupAdminHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPut,
		"/admin/products/"+repository.FirstUUID.String(),
		bytes.NewBufferString(`{
			"name": "",
			"price": 0
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
			"message": "name cannot be blank.; price must be > 0."
		}`,
		res.Body.String(),
	)

	p, exists := repo.FindByID(repository.FirstUUID)
	assert.True(t, exists)
	assert.Equal(t, "MacBook Pro", p.Name)
	assert.Equal(t, 2500.0, p.Price)
}

func TestUpdateProduct_WhenProductNotExists_Returns404(t *testing.T) {
	r, repo := setupAdminHandlerTest(t)
	id, _ := uuid.NewV7()

	req := httptest.NewRequest(
		http.MethodPut,
		"/admin/products/"+id.String(),
		bytes.NewBufferString(`{
			"name": "non-existing-product",
			"price": 1000.1
		}`),
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "PRODUCT_NOT_FOUND",
			"message": "product not found"
		}`,
		res.Body.String(),
	)

	p, exists := repo.FindByID(id)
	assert.False(t, exists)
	assert.Empty(t, p)
}

func TestDeleteProduct_WhenProductExists_DeletesProduct(t *testing.T) {
	r, repo := setupAdminHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/admin/products/"+repository.SecondUUID.String(),
		nil,
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNoContent, res.Code)

	_, exists := repo.FindByID(repository.SecondUUID)
	assert.False(t, exists)
}

func TestDeleteProduct_WhenBadUUID_Returns400(t *testing.T) {
	r, repo := setupAdminHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/admin/products/1234",
		nil,
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INVALID_UUID",
			"message": "invalid UUID"
		}`,
		res.Body.String(),
	)

	products := repo.FindAll()
	assert.Len(t, products, 2)
}
