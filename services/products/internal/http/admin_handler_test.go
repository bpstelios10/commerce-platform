package httpx

import (
	"bytes"
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func setupAdminHandlerTest(t *testing.T) (*chi.Mux, *repository.InMemoryProductRepository) {
	t.Helper()
	repo := repository.NewInMemoryProductRepository()
	adminSvc := service.NewAdminService(repo)
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
			"id": "3",
			"name": "iPad",
			"price": 999
		}`),
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusCreated, res.Code)

	p, exists := repo.FindByID("3")
	assert.True(t, exists)
	assert.Equal(t, "3", p.ID)
	assert.Equal(t, "iPad", p.Name)
	assert.Equal(t, 999.0, p.Price)
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
			"message": "id cannot be blank.; name cannot be blank.; price must be > 0."
		}`,
		res.Body.String(),
	)
}

func TestUpdateProduct_WhenRequestValid_UpdatesProduct(t *testing.T) {
	r, repo := setupAdminHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPut,
		"/admin/products/2",
		bytes.NewBufferString(`{
			"name": "iPhone 15",
			"price": 1500
		}`),
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)

	p, exists := repo.FindByID("2")
	assert.True(t, exists)
	assert.Equal(t, "2", p.ID)
	assert.Equal(t, "iPhone 15", p.Name)
	assert.Equal(t, 1500.0, p.Price)
}

func TestUpdateProduct_WhenBadRequestBody_Returns400(t *testing.T) {
	r, _ := setupAdminHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodPut,
		"/admin/products/1",
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
		"/admin/products/1",
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

	p, exists := repo.FindByID("1")
	assert.True(t, exists)
	assert.Equal(t, "MacBook Pro", p.Name)
	assert.Equal(t, 2500.0, p.Price)
}

func TestDeleteProduct_WhenProductExists_DeletesProduct(t *testing.T) {
	r, repo := setupAdminHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/admin/products/2",
		nil,
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNoContent, res.Code)

	_, exists := repo.FindByID("2")
	assert.False(t, exists)
}
