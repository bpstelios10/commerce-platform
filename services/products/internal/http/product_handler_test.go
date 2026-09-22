package http

import (
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupProductHandlerTest(t *testing.T) (*httptest.Server, *repository.InMemoryProductRepository) {
	t.Helper()
	repo := repository.NewInMemoryProductRepository()
	svc := service.NewProductService(repo)
	handler := NewProductHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	return srv, repo
}

func TestGetProducts_WhenProductsExist_Returns200(t *testing.T) {
	srv, _ := setupProductHandlerTest(t)

	res, err := http.Get(srv.URL + "/products")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var resProducts []map[string]any
	err = json.Unmarshal(body, &resProducts)
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
	srv, _ := setupProductHandlerTest(t)

	res, err := http.Get(srv.URL + "/products/" + repository.FirstUUID.String())

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"id": "`+repository.FirstUUID.String()+`",
			"name": "MacBook Pro",
			"category": "ACCESSORY",
			"price": 2500,
			"stock": 10
		}`,
		string(body),
	)
}

func TestGetProduct_WhenProductNotExists_Returns404(t *testing.T) {
	srv, _ := setupProductHandlerTest(t)
	id, _ := uuid.NewV7()

	res, err := http.Get(srv.URL + "/products/" + id.String())

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "PRODUCT_NOT_FOUND",
			"message": "Product with id [`+id.String()+`] was not found"
		}`,
		string(body),
	)
}

func TestGetProduct_WhenBadUUID_Returns400(t *testing.T) {
	srv, _ := setupProductHandlerTest(t)

	res, err := http.Get(srv.URL + "/products/1234")

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

func TestSearchProducts_WhenOnlyQueryProvided_ReturnsMatches(t *testing.T) {
	srv, _ := setupProductHandlerTest(t)

	res, err := http.Get(srv.URL + "/products/search?query=hoodie")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var products []map[string]any
	err = json.Unmarshal(body, &products)
	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, repository.ThirdUUID.String(), products[0]["id"])
}

func TestSearchProducts_WhenQueryAndMaxPriceProvided_ReturnsCombinedMatches(t *testing.T) {
	srv, _ := setupProductHandlerTest(t)

	res, err := http.Get(srv.URL + "/products/search?query=necklace&maxPrice=200.0")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var products []map[string]any
	err = json.Unmarshal(body, &products)
	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, repository.FourthUUID.String(), products[0]["id"])
}

func TestSearchProducts_WhenOnlyCategoryProvided_ReturnsCategoryMatches(t *testing.T) {
	srv, _ := setupProductHandlerTest(t)

	res, err := http.Get(srv.URL + "/products/search?category=accessory")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var products []map[string]any
	err = json.Unmarshal(body, &products)
	assert.NoError(t, err)
	assert.Len(t, products, 2)
}

func TestSearchProducts_WhenAllCriteriaProvided_ReturnsCombinedMatches(t *testing.T) {
	srv, _ := setupProductHandlerTest(t)

	res, err := http.Get(srv.URL + "/products/search?query=hoodie&maxPrice=100.0&category=clothes")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var products []map[string]any
	err = json.Unmarshal(body, &products)
	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, repository.ThirdUUID.String(), products[0]["id"])
}

func TestSearchProducts_WhenMaxPriceInvalid_Returns400(t *testing.T) {
	srv, _ := setupProductHandlerTest(t)

	res, err := http.Get(srv.URL + "/products/search?maxPrice=abc")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "VALIDATION_ERROR",
			"message": "maxPrice must be a valid number."
		}`,
		string(body),
	)
}
