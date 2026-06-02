package httpx

import (
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestGetProducts_WhenProductsExist_Returns200(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := service.NewProductService(repo)
	handler := NewProductHandler(svc)

	r := chi.NewRouter()

	handler.RegisterRoutes(r)

	req := httptest.NewRequest(
		http.MethodGet,
		"/products",
		nil,
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`[{
			"id": "1",
			"name": "MacBook Pro",
			"price": 2500
		},{
			"id": "2",
			"name": "iPhone",
			"price": 1200
		}]`,
		res.Body.String(),
	)
}

func TestGetProduct_WhenProductExists_Returns200(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := service.NewProductService(repo)
	handler := NewProductHandler(svc)

	r := chi.NewRouter()

	handler.RegisterRoutes(r)

	req := httptest.NewRequest(
		http.MethodGet,
		"/products/1",
		nil,
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"id": "1",
			"name": "MacBook Pro",
			"price": 2500
		}`,
		res.Body.String(),
	)
}

func TestGetProduct_WhenProductNotExists_Returns404(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := service.NewProductService(repo)
	handler := NewProductHandler(svc)

	r := chi.NewRouter()

	handler.RegisterRoutes(r)

	req := httptest.NewRequest(
		http.MethodGet,
		"/products/6",
		nil,
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
}
