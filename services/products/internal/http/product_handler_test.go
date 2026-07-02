package httpx

import (
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

func setupProductHandlerTest(t *testing.T) (*chi.Mux, *repository.InMemoryProductRepository) {
	t.Helper()
	repo := repository.NewInMemoryProductRepository()
	svc := service.NewProductService(repo)
	handler := NewProductHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	return r, repo
}

func TestGetProducts_WhenProductsExist_Returns200(t *testing.T) {
	r, _ := setupProductHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/products",
		nil,
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))

	var resProducts []map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &resProducts)
	assert.NoError(t, err)

	expectedProducts := []map[string]any{
		{
			"id":       repository.FirstUUID.String(),
			"name":     "MacBook Pro",
			"category": "ACCESSORY",
			"price":    2500.0,
			"stock":    float64(10),
		},
		{
			"id":       repository.SecondUUID.String(),
			"name":     "iPhone",
			"category": "ACCESSORY",
			"price":    1200.0,
			"stock":    float64(5),
		},
		{
			"id":       repository.ThirdUUID.String(),
			"name":     "hoodie Mykonos",
			"category": "CLOTHES",
			"price":    80.0,
			"stock":    float64(8),
		},
		{
			"id":       repository.FourthUUID.String(),
			"name":     "Eye necklace",
			"category": "JEWELRY",
			"price":    150.0,
			"stock":    float64(15),
		},
	}

	assert.ElementsMatch(t, expectedProducts, resProducts)
}

func TestGetProduct_WhenProductExists_Returns200(t *testing.T) {
	r, _ := setupProductHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/products/"+repository.FirstUUID.String(),
		nil,
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"id": "`+repository.FirstUUID.String()+`",
			"name": "MacBook Pro",
			"category": "ACCESSORY",
			"price": 2500,
			"stock": 10
		}`,
		res.Body.String(),
	)
}

func TestGetProduct_WhenProductNotExists_Returns404(t *testing.T) {
	r, _ := setupProductHandlerTest(t)
	id, _ := uuid.NewV7()

	req := httptest.NewRequest(
		http.MethodGet,
		"/products/"+id.String(),
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

func TestGetProduct_WhenBadUUID_Returns400(t *testing.T) {
	r, _ := setupProductHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/products/1234",
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
}
